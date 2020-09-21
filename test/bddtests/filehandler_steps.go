/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package bddtests

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/cucumber/godog"
	"github.com/pkg/errors"

	"github.com/trustbloc/fabric-peer-test-common/bddtests"
	"github.com/trustbloc/sidetree-core-go/pkg/document"
	"github.com/trustbloc/sidetree-core-go/pkg/docutil"
	"github.com/trustbloc/sidetree-core-go/pkg/jws"
	"github.com/trustbloc/sidetree-core-go/pkg/patch"
	"github.com/trustbloc/sidetree-core-go/pkg/restapi/helper"
	"github.com/trustbloc/sidetree-core-go/pkg/util/ecsigner"
	"github.com/trustbloc/sidetree-core-go/pkg/util/pubkey"
)

// FileHandlerSteps
type FileHandlerSteps struct {
	httpSteps

	bddContext      *bddtests.BDDContext
	recoveryKey     *ecdsa.PrivateKey
	updateKey       *ecdsa.PrivateKey
	updateKeySigner helper.Signer
	updatePublicKey *jws.JWK
}

// NewFileHandlerSteps
func NewFileHandlerSteps(context *bddtests.BDDContext) *FileHandlerSteps {
	return &FileHandlerSteps{bddContext: context}
}

func (d *FileHandlerSteps) createDocument(url, content, namespace string) error {
	resolved, err := bddtests.ResolveAllVars(content)
	if err != nil {
		return err
	}

	if len(resolved) != 1 {
		return errors.Errorf("expecting 1 var but got %d", len(resolved))
	}

	content = resolved[0]

	opaque, err := d.getOpaqueDocument(content)
	if err != nil {
		return err
	}

	logger.Infof("Creating document at [%s] in namespace [%s] with content %s", url, namespace, opaque)

	req, err := d.getCreateRequest(opaque)
	if err != nil {
		return err
	}

	return d.httpPost(url, req, contentTypeJSON)
}

func (d *FileHandlerSteps) updateDocument(url, docID, jsonPatch string) error {
	logger.Infof("Updating document [%s] at [%s] with patch %s", docID, url, jsonPatch)

	resolvedPatch, err := bddtests.ResolveAllVars(jsonPatch)
	if err != nil {
		return err
	}

	if len(resolvedPatch) != 1 {
		return errors.Errorf("expecting 1 var but got %d", len(resolvedPatch))
	}

	jsonPatch = resolvedPatch[0]

	resolvedDocID, err := bddtests.ResolveAllVars(docID)
	if err != nil {
		return err
	}

	if len(resolvedDocID) != 1 {
		return errors.Errorf("expecting 1 var but got %d", len(resolvedDocID))
	}

	uniqueSuffix := getUniqueSuffix(resolvedDocID[0])

	logger.Infof("Updating document [%s] at [%s] with patch %s", docID, url, jsonPatch)

	req, err := d.getUpdateRequest(uniqueSuffix, jsonPatch)
	if err != nil {
		return err
	}

	return d.httpPost(url, req, contentTypeJSON)
}

func (d *FileHandlerSteps) uploadFile(url, path, contentType string) error {
	logger.Infof("Uploading file [%s] to [%s]", path, url)

	fileBytes := getFile(path)

	req := &UploadFile{
		ContentType: contentType,
		Content:     fileBytes,
	}

	reqBytes, err := json.Marshal(req)
	if err != nil {
		return err
	}

	return d.httpPost(url, reqBytes, contentTypeJSON)
}

func (d *httpSteps) httpGetWithRetryOnNotFound(url string) error {
	return d.httpGetWithRetry(url, 20, http.StatusNotFound)
}

func (d *FileHandlerSteps) getCreateRequest(doc []byte) ([]byte, error) {
	if d.recoveryKey == nil {
		recoveryKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			return nil, err
		}

		updateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			return nil, err
		}

		d.recoveryKey = recoveryKey
		d.updateKey = updateKey
	}

	recoveryCommitment, err := getCommitment(&d.recoveryKey.PublicKey)
	if err != nil {
		return nil, err
	}

	updateCommitment, err := getCommitment(&d.updateKey.PublicKey)
	if err != nil {
		return nil, err
	}

	return helper.NewCreateRequest(&helper.CreateRequestInfo{
		OpaqueDocument:     string(doc),
		RecoveryCommitment: recoveryCommitment,
		UpdateCommitment:   updateCommitment,
		MultihashCode:      sha2_256,
	})
}

func (d *FileHandlerSteps) retrievedFileContains(msg string) error {
	if d.statusCode != http.StatusOK {
		return errors.Errorf("status code: %d - %s", d.statusCode, d.response)
	}

	logger.Infof("check success resp %s contain %s", string(d.response), msg)
	if !strings.Contains(string(d.response), msg) {
		return errors.Errorf("success resp %s doesn't contain %s", d.response, msg)
	}
	return nil
}

func (d *FileHandlerSteps) saveIDToVariable(varName string) error {
	if d.statusCode != http.StatusOK {
		return errors.Errorf("status code: %d - %s", d.statusCode, d.response)
	}

	id := ""
	if err := json.Unmarshal(d.response, &id); err != nil {
		return err
	}

	logger.Infof("Saving ID [%s] to variable [%s]", id, varName)

	bddtests.SetVar(varName, id)
	return nil
}

func (d *FileHandlerSteps) saveDocIDToVariable(varName string) error {
	if d.statusCode != http.StatusOK {
		return errors.Errorf("status code: %d - %s", d.statusCode, d.response)
	}

	var result document.ResolutionResult
	if err := json.Unmarshal(d.response, &result); err != nil {
		return err
	}

	logger.Infof("Got doc %v", result.Document)
	logger.Infof("Saving ID [%s] to variable [%s]", result.Document["id"], varName)

	bddtests.SetVar(varName, result.Document["id"].(string))
	return nil
}

func (d *FileHandlerSteps) setJSONPatchVar(varName, patch string) error {
	var p []interface{}
	err := json.Unmarshal([]byte(patch), &p)
	if err != nil {
		panic(err)
	}

	obj, err := bddtests.ResolveVars(p)
	if err != nil {
		panic(err)
	}

	bytes, err := json.Marshal(obj)
	if err != nil {
		panic(err)
	}

	logger.Infof("Setting variable [%s] to JSON patch %s", varName, bytes)

	bddtests.SetVar(varName, string(bytes))

	return nil
}

func (d *FileHandlerSteps) getOpaqueDocument(content string) ([]byte, error) {
	// generate private key that will be used for document updates and
	// insert public key that correspond to this private key into document (JWK format)
	const updateKeyID = "updateKey"
	if d.updateKeySigner == nil {
		privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			return nil, err
		}

		d.updatePublicKey, err = pubkey.GetPublicKeyJWK(&privateKey.PublicKey)
		if err != nil {
			return nil, err
		}

		// this signer will be used in subsequent update requests
		d.updateKeySigner = ecsigner.New(privateKey, "ES256", updateKeyID)
	}

	doc, err := document.FromBytes([]byte(content))
	if err != nil {
		return nil, err
	}

	return doc.Bytes()
}

func (d *FileHandlerSteps) getUpdateRequest(uniqueSuffix string, jsonPatch string) ([]byte, error) {
	updatePatch, err := patch.NewJSONPatch(jsonPatch)
	if err != nil {
		return nil, err
	}

	updateKey, updateCommitment, err := generateKeyAndCommitment()
	if err != nil {
		return nil, err
	}

	// update key and signer passed in are generated during previous operations
	updatePubKey, err := pubkey.GetPublicKeyJWK(&d.updateKey.PublicKey)
	if err != nil {
		return nil, err
	}

	req, err := helper.NewUpdateRequest(&helper.UpdateRequestInfo{
		DidSuffix:        uniqueSuffix,
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

func getFile(path string) []byte {
	r, err := os.Open(path)
	if err != nil {
		panic(err)
	}

	data, err := ioutil.ReadAll(r)
	if err != nil {
		panic(err)
	}

	return data
}

func getUniqueSuffix(docID string) string {
	pos := strings.LastIndex(docID, docutil.NamespaceDelimiter)
	if pos == -1 {
		return docID
	}

	return docID[pos+1:]
}

// UploadFile contains the file upload request
type UploadFile struct {
	ContentType string `json:"contentType"`
	Content     []byte `json:"content"`
}

// RegisterSteps registers did sidetree steps
func (d *FileHandlerSteps) RegisterSteps(s *godog.Suite) {
	s.Step(`^client sends request to "([^"]*)" to create document with content "([^"]*)" in namespace "([^"]*)"$`, d.createDocument)
	s.Step(`^client sends request to "([^"]*)" to update document "([^"]*)" with patch "([^"]*)"$`, d.updateDocument)
	s.Step(`^client sends request to "([^"]*)" to retrieve file$`, d.httpGetWithRetryOnNotFound)
	s.Step(`^client sends request to "([^"]*)" to upload file "([^"]*)" with content type "([^"]*)"$`, d.uploadFile)
	s.Step(`^the ID of the file is saved to variable "([^"]*)"`, d.saveIDToVariable)
	s.Step(`^the ID of the returned document is saved to variable "([^"]*)"`, d.saveDocIDToVariable)
	s.Step(`^the retrieved file contains "([^"]*)"$`, d.retrievedFileContains)
	s.Step(`^the response has status code (\d+) and error message "([^"]*)"$`, d.checkResponse)
	s.Step(`^variable "([^"]*)" is assigned the JSON patch '([^']*)'$`, d.setJSONPatchVar)
}
