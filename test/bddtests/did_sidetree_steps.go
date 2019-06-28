/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package bddtests

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/trustbloc/fabric-peer-test-common/bddtests"

	"github.com/trustbloc/sidetree-core-go/pkg/document"

	"github.com/trustbloc/sidetree-core-go/pkg/docutil"

	"github.com/DATA-DOG/godog"
	"github.com/go-openapi/swag"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/trustbloc/sidetree-fabric/test/bddtests/restclient"
	"github.com/trustbloc/sidetree-node/models"
)

var logger = logrus.New()

const sha2256 = 18
const didDocNamespace = "did:sidetree:"
const testDocumentURL = "http://localhost:48326/.sidetree/document"

// DIDSideSteps
type DIDSideSteps struct {
	reqEncodedDIDDoc string
	resp             *restclient.HttpRespone
	bddContext       *bddtests.BDDContext
}

// NewDIDSideSteps
func NewDIDSideSteps(context *bddtests.BDDContext) *DIDSideSteps {
	return &DIDSideSteps{bddContext: context}
}

func (d *DIDSideSteps) sendDIDDocument(didDocumentPath, requestType string) error {
	return d.sendDIDDocumentWithID(didDocumentPath, requestType, "")
}

func (d *DIDSideSteps) sendDIDDocumentWithID(didDocumentPath, requestType, didID string) error {
	var err error
	logger.Infof("create did document %s with didID %s", didDocumentPath, didID)
	if didID != "" {
		didID = fmt.Sprintf(`"id": "%s",`, didDocNamespace+didID)
	}

	if requestType == "JSON" {
		req := newCreateRequest(didDocumentPath, didID)
		d.reqEncodedDIDDoc = swag.StringValue(req.Payload)
		d.resp, err = restclient.SendRequest(testDocumentURL, req)
		return err
	}
	d.reqEncodedDIDDoc = encodeDidDocument(didDocumentPath, didID)
	d.resp, err = restclient.SendResolveRequest(testDocumentURL + "/" + didDocNamespace + d.reqEncodedDIDDoc)
	return err
}

func (d *DIDSideSteps) checkErrorResp(errorMsg string) error {
	if !strings.Contains(d.resp.ErrorMsg, errorMsg) {
		return errors.Errorf("error resp %s doesn't contain %s", d.resp.ErrorMsg, errorMsg)
	}
	return nil
}

func (d *DIDSideSteps) checkSuccessResp(msg string) error {
	if d.resp.ErrorMsg != "" {
		return errors.Errorf("error resp %s", d.resp.ErrorMsg)
	}

	if msg == "#didDocumentHash" {
		documentHash, err := docutil.CalculateID(didDocNamespace, d.reqEncodedDIDDoc, sha2256)
		if err != nil {
			return err
		}
		msg = strings.Replace(msg, "#didDocumentHash", documentHash, -1)
	}
	logger.Infof("check success resp %s contain %s", string(d.resp.Payload), msg)
	if !strings.Contains(string(d.resp.Payload), msg) {
		return errors.Errorf("success resp %s doesn't contain %s", d.resp.Payload, msg)
	}
	return nil
}

func (d *DIDSideSteps) resolveDIDDocument() error {
	documentHash, err := docutil.CalculateID(didDocNamespace, d.reqEncodedDIDDoc, sha2256)
	if err != nil {
		return err
	}
	d.resp, err = restclient.SendResolveRequest(testDocumentURL + "/" + documentHash)
	return err
}

func newCreateRequest(didDocumentPath, didID string) *models.Request {
	operation := models.OperationTypeCreate
	alg := "ES256K"
	kid := "#key1"
	payload := encodeDidDocument(didDocumentPath, didID)
	signature := "mAJp4ZHwY5UMA05OEKvoZreRo0XrYe77s3RLyGKArG85IoBULs4cLDBtdpOToCtSZhPvCC2xOUXMGyGXDmmEHg"
	return request(alg, kid, payload, signature, operation)
}

func request(alg, kid, payload, signature string, operation models.OperationType) *models.Request {
	header := &models.Header{
		Alg:       swag.String(alg),
		Kid:       swag.String(kid),
		Operation: operation,
	}
	req := &models.Request{
		Header:    header,
		Payload:   swag.String(payload),
		Signature: swag.String(signature)}
	return req
}

func encodeDidDocument(didDocumentPath, didID string) string {
	r, _ := os.Open(didDocumentPath)
	data, _ := ioutil.ReadAll(r)
	doc, _ := document.FromBytes(data)
	if didID != "" {
		doc["id"] = didID
	}
	// add new key to make the document unique
	doc["unique"] = GenerateUUID()
	bytes, _ := doc.Bytes()
	return base64.URLEncoding.EncodeToString(bytes)
}

// RegisterSteps registers did sidetree steps
func (d *DIDSideSteps) RegisterSteps(s *godog.Suite) {
	s.Step(`^client sends request to create DID document "([^"]*)" as "([^"]*)" with DID id "([^"]*)"$`, d.sendDIDDocumentWithID)
	s.Step(`^check error response contains "([^"]*)"$`, d.checkErrorResp)
	s.Step(`^client sends request to create DID document "([^"]*)" as "([^"]*)"`, d.sendDIDDocument)
	s.Step(`^check success response contains "([^"]*)"$`, d.checkSuccessResp)
	s.Step(`^client sends request to resolve DID document$`, d.resolveDIDDocument)

}
