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

func (d *DIDSideSteps) checkSuccessResp(msg string) error {
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
	logger.Infof("check success resp %s contain %s", string(d.resp.Payload), msg)
	if !strings.Contains(string(d.resp.Payload), msg) {
		return errors.Errorf("success resp %s doesn't contain %s", d.resp.Payload, msg)
	}
	return nil
}

func (d *DIDSideSteps) updateDIDDocument(url, path, value string) error {
	uniqueSuffix, err := d.getUniqueSuffix()
	if err != nil {
		return err
	}

	logger.Infof("update did document: %s", uniqueSuffix)

	req, err := getUpdateRequest(uniqueSuffix, path, value)
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

func getUpdateRequest(did, path, value string) ([]byte, error) {
	return helper.NewUpdateRequest(&helper.UpdateRequestInfo{
		DidUniqueSuffix:   did,
		UpdateRevealValue: []byte(updateOTP),
		Patch:             getUpdatePatch(path, value),
		MultihashCode:     sha2_256,
	})
}

func getUpdatePatch(path, value string) string {
	patchJSON := fmt.Sprintf(`[{"op": "replace", "path":  "%s", "value": "%s"}]`, path, value)

	logger.Infof("JSON Patch: %s", patchJSON)

	return patchJSON
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
	s.Step(`^client sends request to "([^"]*)" to update DID document path "([^"]*)" with value "([^"]*)"$`, d.updateDIDDocument)
	s.Step(`^client sends request to "([^"]*)" to recover DID document "([^"]*)"$`, d.recoverDIDDocument)
	s.Step(`^client sends request to "([^"]*)" to revoke DID document$`, d.revokeDIDDocument)
	s.Step(`^client sends request to "([^"]*)" to resolve DID document with initial value$`, d.resolveDIDDocumentWithInitialValue)
	s.Step(`^check success response contains "([^"]*)"$`, d.checkSuccessResp)
	s.Step(`^client sends request to "([^"]*)" to resolve DID document$`, d.resolveDIDDocument)
}
