/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package bddtests

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
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
	"github.com/trustbloc/sidetree-core-go/pkg/util/ecsigner"
	"github.com/trustbloc/sidetree-core-go/pkg/util/pubkey"

	"github.com/trustbloc/sidetree-fabric/test/bddtests/restclient"
)

var logger = logrus.New()

const (
	sha2_256 = 18

	initialStateParamTemplate = "?-%s-initial-state="

	recoveryOTP = "recoveryOTP"
	updateOTP   = "updateOTP"
)

const addPublicKeysTemplate = `[{
      "id": "%s",
      "type": "JwsVerificationKey2020",
      "usage": ["general"],
      "jwk": {
        	"kty": "EC",
        	"crv": "P-256K",
        	"x": "PUymIqdtF_qxaAqPABSw-C-owT1KYYQbsMKFM-L9fJA",
        	"y": "nM84jDHCMOTGTh_ZdHq4dBBdo4Z5PkEOW9jA8z8IsGc"
      }
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

const docTemplate = `{
  "publicKey": [
	{
  		"id": "%s",
  		"type": "JwsVerificationKey2020",
		"usage": ["ops"],
  		"jwk": %s
	},
   {
     "id": "dual-auth-gen",
     "type": "JwsVerificationKey2020",
     "usage": ["auth", "general"],
     "jwk": %s
   },
   {
     "id": "dual-assertion-gen",
     "type": "Ed25519VerificationKey2018",
     "usage": ["assertion", "general"],
     "jwk": %s
   }
  ],
  "service": [
	{
	   "id": "oidc",
	   "type": "OpenIdConnectVersion1.0Service",
	   "serviceEndpoint": "https://openid.example.com/"
	}, 
	{
	   "id": "hub",
	   "type": "HubService",
	   "serviceEndpoint": "https://hub.example.com/.identity/did:example:0123456789abcdef/"
	}
  ]
}`

// DIDSideSteps
type DIDSideSteps struct {
	createRequest     model.CreateRequest
	reqNamespace      string
	recoveryKeySigner helper.Signer
	updateKeySigner   helper.Signer
	resp              *restclient.HttpResponse
	bddContext        *bddtests.BDDContext
}

// NewDIDSideSteps
func NewDIDSideSteps(context *bddtests.BDDContext) *DIDSideSteps {
	return &DIDSideSteps{bddContext: context}
}

func (d *DIDSideSteps) sendDIDDocument(url, namespace string) error {
	logger.Infof("Creating DID document at %s", url)

	opaqueDoc, err := d.getOpaqueDocument("key1")
	if err != nil {
		return err
	}

	req, err := d.getCreateRequest(opaqueDoc)
	if err != nil {
		return err
	}

	err = json.Unmarshal(req, &d.createRequest)
	if err != nil {
		return err
	}

	d.reqNamespace = namespace

	d.resp, err = restclient.SendRequest(url, req)
	return err
}

func (d *DIDSideSteps) resolveDIDDocumentWithInitialValue(url string) error {
	did, err := d.getDID()
	if err != nil {
		return err
	}

	initialState := d.createRequest.SuffixData + "." + d.createRequest.Delta

	method, err := getMethod(d.reqNamespace)
	if err != nil {
		return err
	}

	initialStateParam := fmt.Sprintf(initialStateParamTemplate, method)

	req := url + "/" + did + initialStateParam + initialState
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

	if d.resp.StatusCode != http.StatusOK {
		return errors.Errorf("request failed with status code %d", d.resp.StatusCode)
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

	var result document.ResolutionResult
	err = json.Unmarshal(d.resp.Payload, &result)
	if err != nil {
		return err
	}

	err = prettyPrint(&result)
	if err != nil {
		return err
	}

	if contains && !strings.Contains(string(d.resp.Payload), msg) {
		return errors.Errorf("success resp %s doesn't contain %s", d.resp.Payload, msg)

	}

	if !contains && strings.Contains(string(d.resp.Payload), msg) {
		return errors.Errorf("success resp %s should NOT contain %s", d.resp.Payload, msg)
	}

	logger.Infof("passed check that success response MUST%s contain %s", action, msg)

	return nil
}

func (d *DIDSideSteps) updateDIDDocument(url string, updatePatch patch.Patch) error {
	uniqueSuffix, err := d.getUniqueSuffix()
	if err != nil {
		return err
	}

	logger.Infof("update did document: %s", uniqueSuffix)

	req, err := d.getUpdateRequest(uniqueSuffix, updatePatch)
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

func (d *DIDSideSteps) deactivateDIDDocument(url string) error {
	uniqueSuffix, err := d.getUniqueSuffix()
	if err != nil {
		return err
	}

	logger.Infof("deactivate did document [%s]from %s", uniqueSuffix, url)

	req, err := d.getDeactivateRequest(uniqueSuffix)
	if err != nil {
		return err
	}

	d.resp, err = restclient.SendRequest(url, req)

	logger.Infof("deactivate status %d, error '%s:'", d.resp.StatusCode, d.resp.ErrorMsg)

	return err
}

func (d *DIDSideSteps) recoverDIDDocument(url string) error {
	uniqueSuffix, err := d.getUniqueSuffix()
	if err != nil {
		return err
	}

	logger.Infof("recover did document [%s]", uniqueSuffix)

	opaqueDoc, err := d.getOpaqueDocument("recoveryKey")
	if err != nil {
		return err
	}

	req, err := d.getRecoverRequest(opaqueDoc, uniqueSuffix)
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
	return docutil.CalculateUniqueSuffix(d.createRequest.SuffixData, sha2_256)
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

func (d *DIDSideSteps) getCreateRequest(doc []byte) ([]byte, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}

	d.recoveryKeySigner = ecsigner.New(privateKey, "ES256", "")
	if err != nil {
		return nil, err
	}

	recoveryPublicKey, err := pubkey.GetPublicKeyJWK(&privateKey.PublicKey)
	if err != nil {
		return nil, err
	}

	return helper.NewCreateRequest(&helper.CreateRequestInfo{
		OpaqueDocument:          string(doc),
		RecoveryKey:             recoveryPublicKey,
		NextRecoveryRevealValue: []byte(recoveryOTP),
		NextUpdateRevealValue:   []byte(updateOTP),
		MultihashCode:           sha2_256,
	})
}

func (d *DIDSideSteps) getRecoverRequest(doc []byte, uniqueSuffix string) ([]byte, error) {
	newPrivateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err

	}

	newRecoveryPublicKey, err := pubkey.GetPublicKeyJWK(&newPrivateKey.PublicKey)
	if err != nil {
		return nil, err
	}

	recoverRequest, err := helper.NewRecoverRequest(&helper.RecoverRequestInfo{
		DidSuffix:               uniqueSuffix,
		OpaqueDocument:          string(doc),
		RecoveryKey:             newRecoveryPublicKey,
		RecoveryRevealValue:     []byte(recoveryOTP),
		NextRecoveryRevealValue: []byte(recoveryOTP),
		NextUpdateRevealValue:   []byte(updateOTP),
		MultihashCode:           sha2_256,
		Signer:                  d.recoveryKeySigner, // sign with old signer
	})

	if err != nil {
		return nil, err
	}

	// update recovery key singer for subsequent requests
	d.recoveryKeySigner = ecsigner.New(newPrivateKey, "ES256", "")

	return recoverRequest, nil
}

func (d *DIDSideSteps) getDeactivateRequest(did string) ([]byte, error) {
	return helper.NewDeactivateRequest(&helper.DeactivateRequestInfo{
		DidSuffix:           did,
		RecoveryRevealValue: []byte(recoveryOTP),
		Signer:              d.recoveryKeySigner,
	})
}

func (d *DIDSideSteps) getUpdateRequest(did string, updatePatch patch.Patch) ([]byte, error) {
	return helper.NewUpdateRequest(&helper.UpdateRequestInfo{
		DidSuffix:             did,
		UpdateRevealValue:     []byte(updateOTP),
		NextUpdateRevealValue: []byte(updateOTP),
		Patch:                 updatePatch,
		MultihashCode:         sha2_256,
		Signer:                d.updateKeySigner,
	})
}

func (d *DIDSideSteps) getOpaqueDocument(keyID string) ([]byte, error) {
	// create operations key (used for document updates)
	opsPrivateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}

	opsPubKey, err := getPubKey(&opsPrivateKey.PublicKey)
	if err != nil {
		return nil, err
	}

	// create general + auth JWS verification key
	jwsPrivateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}

	jwsPubKey, err := getPubKey(&jwsPrivateKey.PublicKey)
	if err != nil {
		return nil, err
	}

	// create general + assertion ed25519 verification key
	ed25519PulicKey, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	ed25519PubKey, err := getPubKey(ed25519PulicKey)
	if err != nil {
		return nil, err
	}

	data := fmt.Sprintf(docTemplate, keyID, opsPubKey, jwsPubKey, ed25519PubKey)

	doc, err := document.FromBytes([]byte(data))
	if err != nil {
		return nil, err
	}

	d.updateKeySigner = ecsigner.New(opsPrivateKey, "ES256", keyID)

	return doc.Bytes()
}

// getMethod returns method from namespace
func getMethod(namespace string) (string, error) {
	const minPartsInNamespace = 2
	parts := strings.Split(namespace, ":")
	if len(parts) < minPartsInNamespace {
		return "", fmt.Errorf("namespace '%s' should have at least two parts", namespace)
	}

	return parts[1], nil
}

func getPubKey(pubKey interface{}) (string, error) {
	publicKey, err := pubkey.GetPublicKeyJWK(pubKey)
	if err != nil {
		return "", err
	}

	opsPubKeyBytes, err := json.Marshal(publicKey)
	if err != nil {
		return "", err
	}

	return string(opsPubKeyBytes), nil
}

func prettyPrint(result *document.ResolutionResult) error {
	b, err := json.MarshalIndent(result, "", " ")
	if err != nil {
		return err
	}

	logger.Infof(string(b))

	return nil
}

// RegisterSteps registers did sidetree steps
func (d *DIDSideSteps) RegisterSteps(s *godog.Suite) {
	s.Step(`^check error response contains "([^"]*)"$`, d.checkErrorResp)
	s.Step(`^client sends request to "([^"]*)" to create DID document in namespace "([^"]*)"$`, d.sendDIDDocument)
	s.Step(`^client sends request to "([^"]*)" to update DID document path "([^"]*)" with value "([^"]*)"$`, d.updateDIDDocumentWithJSONPatch)
	s.Step(`^client sends request to "([^"]*)" to add public key with ID "([^"]*)" to DID document$`, d.addPublicKeyToDIDDocument)
	s.Step(`^client sends request to "([^"]*)" to remove public key with ID "([^"]*)" from DID document$`, d.removePublicKeyFromDIDDocument)
	s.Step(`^client sends request to "([^"]*)" to add service endpoint with ID "([^"]*)" to DID document$`, d.addServiceEndpointToDIDDocument)
	s.Step(`^client sends request to "([^"]*)" to remove service endpoint with ID "([^"]*)" from DID document$`, d.removeServiceEndpointsFromDIDDocument)
	s.Step(`^client sends request to "([^"]*)" to recover DID document$`, d.recoverDIDDocument)
	s.Step(`^client sends request to "([^"]*)" to deactivate DID document$`, d.deactivateDIDDocument)
	s.Step(`^client sends request to "([^"]*)" to resolve DID document with initial value$`, d.resolveDIDDocumentWithInitialValue)
	s.Step(`^check success response contains "([^"]*)"$`, d.checkSuccessRespContains)
	s.Step(`^check success response does NOT contain "([^"]*)"$`, d.checkSuccessRespDoesNotContain)
	s.Step(`^client sends request to "([^"]*)" to resolve DID document$`, d.resolveDIDDocument)
}
