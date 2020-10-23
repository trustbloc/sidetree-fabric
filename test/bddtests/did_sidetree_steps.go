/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package bddtests

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/json"
	"fmt"
	mrand "math/rand"
	"net/http"
	"strings"

	"github.com/cucumber/godog"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/trustbloc/fabric-peer-test-common/bddtests"
	"github.com/trustbloc/sidetree-core-go/pkg/canonicalizer"
	"github.com/trustbloc/sidetree-core-go/pkg/commitment"
	"github.com/trustbloc/sidetree-core-go/pkg/document"
	"github.com/trustbloc/sidetree-core-go/pkg/docutil"
	"github.com/trustbloc/sidetree-core-go/pkg/patch"
	"github.com/trustbloc/sidetree-core-go/pkg/versions/0_1/client"
	"github.com/trustbloc/sidetree-core-go/pkg/util/ecsigner"
	"github.com/trustbloc/sidetree-core-go/pkg/util/pubkey"
	"github.com/trustbloc/sidetree-core-go/pkg/versions/0_1/model"
)

var logger = logrus.New()

const (
	sha2_256 = 18

	initialStateSeparator = ":"
)

const addPublicKeysTemplate = `[{
      "id": "%s",
      "type": "JwsVerificationKey2020",
      "purpose": ["general"],
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
		"endpoint": "http://hub.my-personal-server.com"
    }
  ]`

const removeServicesTemplate = `["%s"]`

const docTemplate = `{
  "publicKey": [
   {
     "id": "%s",
     "type": "JwsVerificationKey2020",
     "purpose": ["auth", "general"],
     "jwk": %s
   },
   {
     "id": "dual-assertion-gen",
     "type": "Ed25519VerificationKey2018",
     "purpose": ["assertion", "general"],
     "jwk": %s
   }
  ],
  "service": [
	{
	   "id": "oidc",
	   "type": "OpenIdConnectVersion1.0Service",
	   "endpoint": "https://openid.example.com/"
	}, 
	{
	   "id": "hub",
	   "type": "HubService",
	   "endpoint": "https://hub.example.com/.identity/did:example:0123456789abcdef/"
	}
  ]
}`

// DIDSideSteps
type DIDSideSteps struct {
	httpSteps

	createRequest *model.CreateRequest
	reqNamespace  string
	recoveryKey   *ecdsa.PrivateKey
	updateKey     *ecdsa.PrivateKey
	bddContext    *bddtests.BDDContext
	dids          []string
}

// NewDIDSideSteps
func NewDIDSideSteps(context *bddtests.BDDContext) *DIDSideSteps {
	return &DIDSideSteps{bddContext: context}
}

func (d *DIDSideSteps) sendDIDDocument(url, namespace string) error {
	logger.Infof("Creating DID document at %s", url)

	opaqueDoc, err := getOpaqueDocument("createKey")
	if err != nil {
		return err
	}

	reqBytes, err := d.getCreateRequest(opaqueDoc)
	if err != nil {
		return err
	}

	var req model.CreateRequest
	err = json.Unmarshal(reqBytes, &req)
	if err != nil {
		return err
	}

	d.createRequest = &req
	d.reqNamespace = namespace

	return d.httpPost(url, reqBytes, contentTypeJSON)
}

func (d *DIDSideSteps) resolveDIDDocumentWithInitialValue(url string) error {
	did, err := d.getDID()
	if err != nil {
		return err
	}

	initialState, err := d.getInitialState()
	if err != nil {
		return err
	}

	req := url + "/" + did + initialStateSeparator + initialState

	return d.httpGet(req)
}

func (d *DIDSideSteps) getInitialState() (string, error) {
	createReq := &model.CreateRequest{
		Delta:      d.createRequest.Delta,
		SuffixData: d.createRequest.SuffixData,
	}

	bytes, err := canonicalizer.MarshalCanonical(createReq)
	if err != nil {
		return "", err
	}

	return docutil.EncodeToString(bytes), nil
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

	if d.statusCode != http.StatusOK {
		return errors.Errorf("request failed with status code %d", d.statusCode)
	}

	if msg == "#didDocumentHash" {
		msg = strings.Replace(msg, "#didDocumentHash", documentHash, -1)
	}

	action := " "
	if !contains {
		action = " NOT"
	}

	var result document.ResolutionResult
	err = json.Unmarshal(d.response, &result)
	if err != nil {
		return err
	}

	err = prettyPrint(&result)
	if err != nil {
		return err
	}

	if contains && !strings.Contains(string(d.response), msg) {
		return errors.Errorf("success resp %s doesn't contain %s", d.response, msg)

	}

	if !contains && strings.Contains(string(d.response), msg) {
		return errors.Errorf("success resp %s should NOT contain %s", d.response, msg)
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

	return d.httpPost(url, req, contentTypeJSON)
}

func (d *DIDSideSteps) resolveDIDDocument(url string) error {
	documentHash, err := d.getDID()
	if err != nil {
		return err
	}

	logger.Infof("Resolving DID document %s from %s", documentHash, url)

	return d.httpGetWithRetry(url+"/"+documentHash, 20, http.StatusNotFound)
}

func (d *DIDSideSteps) resolveDIDDocumentWithAlias(url, alias string) error {
	uniqueSuffix, err := d.getUniqueSuffix()
	if err != nil {
		return err
	}

	did := alias + docutil.NamespaceDelimiter + uniqueSuffix

	logger.Infof("Resolving DID document %s from %s", did, url)

	return d.httpGetWithRetry(url+"/"+did, 20, http.StatusNotFound)
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

	return d.httpPost(url, req, contentTypeJSON)
}

func (d *DIDSideSteps) recoverDIDDocument(url string) error {
	uniqueSuffix, err := d.getUniqueSuffix()
	if err != nil {
		return err
	}

	logger.Infof("recover did document [%s]", uniqueSuffix)

	opaqueDoc, err := getOpaqueDocument("recoveryKey")
	if err != nil {
		return err
	}

	req, err := d.getRecoverRequest(opaqueDoc, uniqueSuffix)
	if err != nil {
		return err
	}

	return d.httpPost(url, req, contentTypeJSON)
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
	return docutil.CalculateModelMultihash(d.createRequest.SuffixData, sha2_256)
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
	data, recoveryKey, updateKey, err := getCreateRequest(doc)
	if err != nil {
		return nil, err
	}

	d.recoveryKey = recoveryKey
	d.updateKey = updateKey

	return data, nil
}

func getCreateRequest(doc []byte) ([]byte, *ecdsa.PrivateKey, *ecdsa.PrivateKey, error) {
	recoveryKey, recoveryCommitment, err := generateKeyAndCommitment()
	if err != nil {
		return nil, nil, nil, err
	}

	updateKey, updateCommitment, err := generateKeyAndCommitment()
	if err != nil {
		return nil, nil, nil, err
	}

	data, err := client.NewCreateRequest(&client.CreateRequestInfo{
		OpaqueDocument:     string(doc),
		RecoveryCommitment: recoveryCommitment,
		UpdateCommitment:   updateCommitment,
		MultihashCode:      sha2_256,
	})
	if err != nil {
		return nil, nil, nil, err
	}

	return data, recoveryKey, updateKey, nil
}

func generateKeyAndCommitment() (*ecdsa.PrivateKey, string, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, "", err
	}

	c, err := getCommitment(&key.PublicKey)
	if err != nil {
		return nil, "", err
	}

	return key, c, nil
}

func getCommitment(key *ecdsa.PublicKey) (string, error) {
	pubKey, err := pubkey.GetPublicKeyJWK(key)
	if err != nil {
		return "", err
	}

	return commitment.Calculate(pubKey, sha2_256, crypto.SHA256)
}

func (d *DIDSideSteps) getRecoverRequest(doc []byte, uniqueSuffix string) ([]byte, error) {
	recoveryKey, recoveryCommitment, err := generateKeyAndCommitment()
	if err != nil {
		return nil, err
	}

	updateKey, updateCommitment, err := generateKeyAndCommitment()
	if err != nil {
		return nil, err
	}

	// recovery key and signer passed in are generated during previous operations
	recoveryPubKey, err := pubkey.GetPublicKeyJWK(&d.recoveryKey.PublicKey)
	if err != nil {
		return nil, err
	}

	recoverRequest, err := client.NewRecoverRequest(&client.RecoverRequestInfo{
		DidSuffix:          uniqueSuffix,
		OpaqueDocument:     string(doc),
		RecoveryKey:        recoveryPubKey,
		RecoveryCommitment: recoveryCommitment,
		UpdateCommitment:   updateCommitment,
		MultihashCode:      sha2_256,
		Signer:             ecsigner.New(d.recoveryKey, "ES256", ""), // sign with old signer
	})

	if err != nil {
		return nil, err
	}

	// update recovery and update key for subsequent requests
	d.recoveryKey = recoveryKey
	d.updateKey = updateKey

	return recoverRequest, nil
}

func (d *DIDSideSteps) getDeactivateRequest(did string) ([]byte, error) {
	// recovery key and signer passed in are generated during previous operations
	recoveryPubKey, err := pubkey.GetPublicKeyJWK(&d.recoveryKey.PublicKey)
	if err != nil {
		return nil, err
	}

	return client.NewDeactivateRequest(&client.DeactivateRequestInfo{
		DidSuffix:   did,
		RecoveryKey: recoveryPubKey,
		Signer:      ecsigner.New(d.recoveryKey, "ES256", ""),
	})
}

func (d *DIDSideSteps) getUpdateRequest(did string, updatePatch patch.Patch) ([]byte, error) {
	updateKey, updateCommitment, err := generateKeyAndCommitment()
	if err != nil {
		return nil, err
	}

	// update key and signer passed in are generated during previous operations
	updatePubKey, err := pubkey.GetPublicKeyJWK(&d.updateKey.PublicKey)
	if err != nil {
		return nil, err
	}

	req, err := client.NewUpdateRequest(&client.UpdateRequestInfo{
		DidSuffix:        did,
		UpdateCommitment: updateCommitment,
		UpdateKey:        updatePubKey,
		Patches:          []patch.Patch{updatePatch},
		MultihashCode:    sha2_256,
		Signer:           ecsigner.New(d.updateKey, "ES256", "update-kid"),
	})

	if err != nil {
		return nil, err
	}

	// update update key for subsequent update requests
	d.updateKey = updateKey

	return req, nil
}

func getOpaqueDocument(keyID string) ([]byte, error) {
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

	data := fmt.Sprintf(docTemplate, keyID, jwsPubKey, ed25519PubKey)

	doc, err := document.FromBytes([]byte(data))
	if err != nil {
		return nil, err
	}

	return doc.Bytes()
}

func (d *DIDSideSteps) httpPost(url string, req []byte, contentType string) error {
	resp, statusCode, header, err := bddtests.HTTPPost(url, req, contentType)
	if err != nil {
		return err
	}

	d.setResponse(statusCode, resp, header)

	return nil
}

func (d *DIDSideSteps) httpGet(url string) error {
	resp, statusCode, header, err := bddtests.HTTPGet(url)
	if err != nil {
		return err
	}

	d.setResponse(statusCode, resp, header)

	return nil
}

func (d *DIDSideSteps) setResponse(statusCode int, response []byte, header http.Header) {
	d.statusCode = statusCode
	d.response = response

	logger.Infof("Got header: %s", header)

	contentType, ok := header["Content-Type"]
	if ok {
		d.contentType = contentType[0]
	}
}

func (d *DIDSideSteps) createDIDDocuments(strURLs string, num int, concurrency int) error {
	err := bddtests.ResolveVarsInExpression(&strURLs)
	if err != nil {
		return err
	}

	logger.Infof("Creating %d DID document(s) at %s using a concurrency of %d", num, strURLs, concurrency)

	urls := strings.Split(strURLs, ",")

	p := NewWorkerPool(concurrency)

	p.Start()

	for i := 0; i < num; i++ {
		p.Submit(&createDIDRequest{
			url: urls[mrand.Intn(len(urls))],
		})
	}

	p.Stop()

	logger.Infof("Got %d responses for %d requests", len(p.responses), num)

	if len(p.responses) != num {
		return errors.Errorf("expecting %d responses but got %d", num, len(p.responses))
	}

	for _, resp := range p.responses {
		req := resp.Request.(*createDIDRequest)
		if resp.Err != nil {
			logger.Infof("Got error from [%s]: %s", req.url, resp.Err)
			return resp.Err
		}

		did := resp.Resp.(string)
		logger.Infof("Got DID from [%s]: %s", req.url, did)
		d.dids = append(d.dids, did)
	}

	return nil
}

func (d *DIDSideSteps) verifyDIDDocuments(strURLs string) error {
	err := bddtests.ResolveVarsInExpression(&strURLs)
	if err != nil {
		return err
	}

	logger.Infof("Verifying the %d DID document(s) that were created", len(d.dids))

	urls := strings.Split(strURLs, ",")

	for _, did := range d.dids {
		if err := d.verifyDID(urls[mrand.Intn(len(urls))], did); err != nil {
			return err
		}
	}

	d.dids = nil

	return nil
}

func (d *DIDSideSteps) verifyDID(url, did string) error {
	logger.Infof("Verifying DID %s from %s", did, url)

	err := d.httpGetWithRetry(url+"/"+did, 20, http.StatusNotFound)
	if err != nil {
		return errors.WithMessagef(err, "failed to resolve DID [%s]", did)
	}

	if d.statusCode != http.StatusOK {
		return errors.Errorf("failed to resolve DID [%s] - Status code %d: %s", did, d.statusCode, d.response)
	}

	logger.Infof(".. successfully verified DID %s from %s", did, url)

	return nil
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

type createDIDRequest struct {
	url string
}

func (r *createDIDRequest) Invoke() (interface{}, error) {
	logger.Infof("Creating DID document at %s", r.url)

	opaqueDoc, err := getOpaqueDocument("key1")
	if err != nil {
		return "", err
	}

	req, _, _, err := getCreateRequest(opaqueDoc)
	if err != nil {
		return "", err
	}

	client := bddtests.HTTPClient{}
	docBytes, _, _, err := client.Post(r.url, req, contentTypeJSON)
	if err != nil {
		return "", err
	}

	logger.Infof("... got DID document: %s", docBytes)

	var doc document.DIDDocument
	err = json.Unmarshal(docBytes, &doc)
	if err != nil {
		return "", err
	}

	didDocument, ok := doc["didDocument"]
	if !ok {
		return "", errors.Errorf("Response is missing field 'didDocument': %s", docBytes)
	}

	return didDocument.(map[string]interface{})["id"].(string), nil
}

// RegisterSteps registers did sidetree steps
func (d *DIDSideSteps) RegisterSteps(s *godog.Suite) {
	s.Step(`^check error response contains "([^"]*)"$`, d.checkErrorResponse)
	s.Step(`^client sends request to "([^"]*)" to create DID document in namespace "([^"]*)"$`, d.sendDIDDocument)
	s.Step(`^client sends request to "([^"]*)" to create (\d+) DID documents using (\d+) concurrent requests$`, d.createDIDDocuments)
	s.Step(`^client sends request to "([^"]*)" to verify the DID documents that were created$`, d.verifyDIDDocuments)
	s.Step(`^client sends request to "([^"]*)" to update DID document path "([^"]*)" with value "([^"]*)"$`, d.updateDIDDocumentWithJSONPatch)
	s.Step(`^client sends request to "([^"]*)" to add public key with ID "([^"]*)" to DID document$`, d.addPublicKeyToDIDDocument)
	s.Step(`^client sends request to "([^"]*)" to remove public key with ID "([^"]*)" from DID document$`, d.removePublicKeyFromDIDDocument)
	s.Step(`^client sends request to "([^"]*)" to add service endpoint with ID "([^"]*)" to DID document$`, d.addServiceEndpointToDIDDocument)
	s.Step(`^client sends request to "([^"]*)" to remove service endpoint with ID "([^"]*)" from DID document$`, d.removeServiceEndpointsFromDIDDocument)
	s.Step(`^client sends request to "([^"]*)" to recover DID document$`, d.recoverDIDDocument)
	s.Step(`^client sends request to "([^"]*)" to deactivate DID document$`, d.deactivateDIDDocument)
	s.Step(`^client sends request to "([^"]*)" to resolve DID document with initial state$`, d.resolveDIDDocumentWithInitialValue)
	s.Step(`^check success response contains "([^"]*)"$`, d.checkSuccessRespContains)
	s.Step(`^check success response does NOT contain "([^"]*)"$`, d.checkSuccessRespDoesNotContain)
	s.Step(`^client sends request to "([^"]*)" to resolve DID document$`, d.resolveDIDDocument)
	s.Step(`^client sends request to "([^"]*)" to resolve DID document with alias "([^"]*)"$`, d.resolveDIDDocumentWithAlias)
}
