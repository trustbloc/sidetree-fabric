/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package bddtests

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/cucumber/godog"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/trustbloc/fabric-peer-test-common/bddtests"
	"github.com/trustbloc/sidetree-core-go/pkg/document"
	"github.com/trustbloc/sidetree-core-go/pkg/docutil"
	"github.com/trustbloc/sidetree-core-go/pkg/patch"
	"github.com/trustbloc/sidetree-core-go/pkg/restapi/helper"
	"github.com/trustbloc/sidetree-core-go/pkg/restapi/model"

	"github.com/trustbloc/sidetree-fabric/test/bddtests/restclient"
)

var logger = logrus.New()

const (
	sha2_256           = 18
	initialValuesParam = ";initial-values="

	recoveryOTP = "recoveryOTP"
	updateOTP   = "updateOTP"
)

const addPublicKeysTemplate = `[
	{
      "id": "%s",
      "usage": ["ops"],
      "type": "Secp256k1VerificationKey2018",
      "publicKeyHex": "02b97c30de767f084ce3080168ee293053ba33b235d7116a3263d29f1450936b71"
    }
  ]`

const removePublicKeysTemplate = `["%s"]`

const addServicesTemplate = `[
    {
      "id": "%s",
      "type": "SecureDataStore",
      "serviceEndpoint": "http://hub.my-personal-server.com"
    }
  ]`

const removeServicesTemplate = `["%s"]`

// DIDSideSteps
type DIDSideSteps struct {
	createRequest []byte
	reqNamespace  string
	resp          *restclient.HttpResponse
	bddContext    *bddtests.BDDContext
}

// NewDIDSideSteps
func NewDIDSideSteps(context *bddtests.BDDContext) *DIDSideSteps {
	return &DIDSideSteps{bddContext: context}
}

func (d *DIDSideSteps) sendDIDDocument(url, didDocumentPath, namespace string) error {
	logger.Infof("Creating DID document at %s", url)

	opaqueDoc := getOpaqueDocument(didDocumentPath)
	req, err := getCreateRequest(opaqueDoc)
	if err != nil {
		return err
	}

	d.createRequest = req
	d.reqNamespace = namespace

	d.resp, err = restclient.SendRequest(url, req)
	return err
}

func (d *DIDSideSteps) resolveDIDDocumentWithInitialValue(url string) error {
	did, err := d.getDID()
	if err != nil {
		return err
	}

	req := url + "/" + did + initialValuesParam + docutil.EncodeToString(d.createRequest)
	logger.Infof("Sending request: %s", req)
	d.resp, err = restclient.SendResolveRequest(req)
	logger.Infof("... got response: %s", d.resp.Payload)
	return err
}

func (d *DIDSideSteps) checkErrorResp(errorMsg string) error {
	if !strings.Contains(d.resp.ErrorMsg, errorMsg) {
		return errors.Errorf("error resp %s doesn't contain %s", d.resp.ErrorMsg, errorMsg)
	}
	return nil
}

func (d *DIDSideSteps) checkSuccessRespContains(msg string) error {
	return d.checkSuccessResp(msg, true)
}

func (d *DIDSideSteps) checkSuccessRespDoesNotContain(msg string) error {
	return d.checkSuccessResp(msg, false)
}

func (d *DIDSideSteps) checkSuccessResp(msg string, contains bool) error {
	documentHash, err := d.getDID()
	if err != nil {
		return err
	}

	if d.resp.ErrorMsg != "" {
		return errors.Errorf("error resp: [%s] - DID ID [%s]", d.resp.ErrorMsg, documentHash)
	}

	if msg == "#didDocumentHash" {
		msg = strings.Replace(msg, "#didDocumentHash", documentHash, -1)
	}

	action := " "
	if !contains {
		action = " NOT"
	}

	logger.Infof("check success resp %s MUST%s contain %s", string(d.resp.Payload), action, msg)
	if contains && !strings.Contains(string(d.resp.Payload), msg) {
		return errors.Errorf("success resp %s doesn't contain %s", d.resp.Payload, msg)
	}

	if !contains && strings.Contains(string(d.resp.Payload), msg) {
		return errors.Errorf("success resp %s should NOT contain %s", d.resp.Payload, msg)
	}
	return nil
}

func (d *DIDSideSteps) updateDIDDocument(url string, updatePatch patch.Patch) error {
	uniqueSuffix, err := d.getUniqueSuffix()
	if err != nil {
		return err
	}

	logger.Infof("update did document: %s", uniqueSuffix)

	req, err := getUpdateRequest(uniqueSuffix, updatePatch)
	if err != nil {
		return err
	}

	d.resp, err = restclient.SendRequest(url, req)
	return err
}

func (d *DIDSideSteps) resolveDIDDocument(url string) error {
	documentHash, err := d.getDID()
	if err != nil {
		return err
	}

	logger.Infof("Resolving DID document %s from %s", documentHash, url)

	remainingAttempts := 20
	for {
		d.resp, err = restclient.SendResolveRequest(url + "/" + documentHash)
		if err != nil {
			return err
		}
		if d.resp.StatusCode == http.StatusNotFound {
			logger.Infof("Document not found: %s. Remaining attempts: %d", documentHash, remainingAttempts)
			remainingAttempts--
			if remainingAttempts > 0 {
				time.Sleep(time.Second)
				continue
			}
		}
		return nil
	}
}

func (d *DIDSideSteps) revokeDIDDocument(url string) error {
	uniqueSuffix, err := d.getUniqueSuffix()
	if err != nil {
		return err
	}

	logger.Infof("revoke did document [%s]from %s", uniqueSuffix, url)

	req, err := getRevokeRequest(uniqueSuffix)
	if err != nil {
		return err
	}

	d.resp, err = restclient.SendRequest(url, req)

	logger.Infof("revoke status %d, error '%s:'", d.resp.StatusCode, d.resp.ErrorMsg)

	return err
}

func (d *DIDSideSteps) recoverDIDDocument(url, didDocumentPath string) error {
	uniqueSuffix, err := d.getUniqueSuffix()
	if err != nil {
		return err
	}

	logger.Infof("revoke did document [%s]from %s", uniqueSuffix, didDocumentPath)

	opaqueDoc := getOpaqueDocument(didDocumentPath)
	req, err := getRecoverRequest(opaqueDoc, uniqueSuffix)
	if err != nil {
		return err
	}

	d.resp, err = restclient.SendRequest(url, req)
	return err
}

func (d *DIDSideSteps) getDID() (string, error) {
	uniqueSuffix, err := d.getUniqueSuffix()
	if err != nil {
		return "", err
	}

	didID := d.reqNamespace + docutil.NamespaceDelimiter + uniqueSuffix
	return didID, nil
}

func (d *DIDSideSteps) getUniqueSuffix() (string, error) {
	var createReq model.CreateRequest
	err := json.Unmarshal(d.createRequest, &createReq)
	if err != nil {
		return "", err
	}

	return docutil.CalculateUniqueSuffix(createReq.SuffixData, sha2_256)
}

func (d *DIDSideSteps) updateDIDDocumentWithJSONPatch(url, path, value string) error {
	updatePatch, err := getJSONPatch(path, value)
	if err != nil {
		return err
	}

	return d.updateDIDDocument(url, updatePatch)
}

func (d *DIDSideSteps) addPublicKeyToDIDDocument(url, keyID string) error {
	updatePatch, err := getAddPublicKeysPatch(keyID)
	if err != nil {
		return err
	}

	return d.updateDIDDocument(url, updatePatch)
}

func (d *DIDSideSteps) removePublicKeyFromDIDDocument(url, keyID string) error {
	updatePatch, err := getRemovePublicKeysPatch(keyID)
	if err != nil {
		return err
	}

	return d.updateDIDDocument(url, updatePatch)
}

func (d *DIDSideSteps) addServiceEndpointToDIDDocument(url, keyID string) error {
	updatePatch, err := getAddServiceEndpointsPatch(keyID)
	if err != nil {
		return err
	}

	return d.updateDIDDocument(url, updatePatch)
}

func (d *DIDSideSteps) removeServiceEndpointsFromDIDDocument(url, keyID string) error {
	updatePatch, err := getRemoveServiceEndpointsPatch(keyID)
	if err != nil {
		return err
	}

	return d.updateDIDDocument(url, updatePatch)
}

func getJSONPatch(path, value string) (patch.Patch, error) {
	patchJSON := fmt.Sprintf(`[{"op": "replace", "path":  "%s", "value": "%s"}]`, path, value)
	logger.Infof("creating JSON patch: %s", patchJSON)
	return patch.NewJSONPatch(patchJSON)
}

func getAddPublicKeysPatch(keyID string) (patch.Patch, error) {
	addPubKeys := fmt.Sprintf(addPublicKeysTemplate, keyID)
	logger.Infof("creating add public keys patch: %s", addPubKeys)
	return patch.NewAddPublicKeysPatch(addPubKeys)
}

func getRemovePublicKeysPatch(keyID string) (patch.Patch, error) {
	removePubKeys := fmt.Sprintf(removePublicKeysTemplate, keyID)
	logger.Infof("creating remove public keys patch: %s", removePubKeys)
	return patch.NewRemovePublicKeysPatch(removePubKeys)
}

func getAddServiceEndpointsPatch(svcID string) (patch.Patch, error) {
	addServices := fmt.Sprintf(addServicesTemplate, svcID)
	logger.Infof("creating add service endpoints patch: %s", addServices)
	return patch.NewAddServiceEndpointsPatch(addServices)
}

func getRemoveServiceEndpointsPatch(keyID string) (patch.Patch, error) {
	removeServices := fmt.Sprintf(removeServicesTemplate, keyID)
	logger.Infof("creating remove service endpoints patch: %s", removeServices)
	return patch.NewRemoveServiceEndpointsPatch(removeServices)
}

func getCreateRequest(doc string) ([]byte, error) {
	return helper.NewCreateRequest(&helper.CreateRequestInfo{
		OpaqueDocument:          doc,
		RecoveryKey:             "recoveryKey",
		NextRecoveryRevealValue: []byte(recoveryOTP),
		NextUpdateRevealValue:   []byte(updateOTP),
		MultihashCode:           sha2_256,
	})
}

func getRecoverRequest(doc, uniqueSuffix string) ([]byte, error) {
	return helper.NewRecoverRequest(&helper.RecoverRequestInfo{
		DidUniqueSuffix:         uniqueSuffix,
		OpaqueDocument:          doc,
		RecoveryKey:             "HEX",
		RecoveryRevealValue:     []byte(recoveryOTP),
		NextRecoveryRevealValue: []byte(recoveryOTP),
		NextUpdateRevealValue:   []byte(updateOTP),
		MultihashCode:           sha2_256,
	})
}

func getRevokeRequest(did string) ([]byte, error) {
	return helper.NewRevokeRequest(&helper.RevokeRequestInfo{
		DidUniqueSuffix:     did,
		RecoveryRevealValue: []byte(recoveryOTP),
	})
}

func getUpdateRequest(did string, updatePatch patch.Patch) ([]byte, error) {
	return helper.NewUpdateRequest(&helper.UpdateRequestInfo{
		DidUniqueSuffix:       did,
		UpdateRevealValue:     []byte(updateOTP),
		NextUpdateRevealValue: []byte(updateOTP),
		Patch:                 updatePatch,
		MultihashCode:         sha2_256,
	})
}

func getOpaqueDocument(didDocumentPath string) string {
	r, _ := os.Open(didDocumentPath)
	data, _ := ioutil.ReadAll(r)
	doc, _ := document.FromBytes(data)

	// add new key to make the document unique
	doc["unique"] = GenerateUUID()
	bytes, _ := doc.Bytes()
	return string(bytes)
}

// RegisterSteps registers did sidetree steps
func (d *DIDSideSteps) RegisterSteps(s *godog.Suite) {
	s.Step(`^check error response contains "([^"]*)"$`, d.checkErrorResp)
	s.Step(`^client sends request to "([^"]*)" to create DID document "([^"]*)" in namespace "([^"]*)"$`, d.sendDIDDocument)
	s.Step(`^client sends request to "([^"]*)" to update DID document path "([^"]*)" with value "([^"]*)"$`, d.updateDIDDocumentWithJSONPatch)
	s.Step(`^client sends request to "([^"]*)" to add public key with ID "([^"]*)" to DID document$`, d.addPublicKeyToDIDDocument)
	s.Step(`^client sends request to "([^"]*)" to remove public key with ID "([^"]*)" from DID document$`, d.removePublicKeyFromDIDDocument)
	s.Step(`^client sends request to "([^"]*)" to add service endpoint with ID "([^"]*)" to DID document$`, d.addServiceEndpointToDIDDocument)
	s.Step(`^client sends request to "([^"]*)" to remove service endpoint with ID "([^"]*)" from DID document$`, d.removeServiceEndpointsFromDIDDocument)
	s.Step(`^client sends request to "([^"]*)" to recover DID document "([^"]*)"$`, d.recoverDIDDocument)
	s.Step(`^client sends request to "([^"]*)" to revoke DID document$`, d.revokeDIDDocument)
	s.Step(`^client sends request to "([^"]*)" to resolve DID document with initial value$`, d.resolveDIDDocumentWithInitialValue)
	s.Step(`^check success response contains "([^"]*)"$`, d.checkSuccessRespContains)
	s.Step(`^check success response does NOT contain "([^"]*)"$`, d.checkSuccessRespDoesNotContain)
	s.Step(`^client sends request to "([^"]*)" to resolve DID document$`, d.resolveDIDDocument)
}
