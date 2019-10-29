/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package bddtests

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/DATA-DOG/godog"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/trustbloc/fabric-peer-test-common/bddtests"
	"github.com/trustbloc/sidetree-core-go/pkg/document"
	"github.com/trustbloc/sidetree-core-go/pkg/docutil"
	"github.com/trustbloc/sidetree-core-go/pkg/restapi/model"
	"github.com/trustbloc/sidetree-fabric/test/bddtests/restclient"
)

var logger = logrus.New()

const sha2256 = 18
const didDocNamespace = "did:sidetree:"
const testDocumentURL = "http://localhost:48326/document"

// DIDSideSteps
type DIDSideSteps struct {
	reqEncodedDIDDoc string
	resp             *restclient.HttpResponse
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
		d.reqEncodedDIDDoc = req.Payload
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
	documentHash, err := docutil.CalculateID(didDocNamespace, d.reqEncodedDIDDoc, sha2256)
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

func (d *DIDSideSteps) resolveDIDDocument() error {
	documentHash, err := docutil.CalculateID(didDocNamespace, d.reqEncodedDIDDoc, sha2256)
	if err != nil {
		return err
	}

	remainingAttempts := 20
	for {
		d.resp, err = restclient.SendResolveRequest(testDocumentURL + "/" + documentHash)
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

func newCreateRequest(didDocumentPath, didID string) *model.Request {
	operation := model.OperationTypeCreate
	alg := "ES256K"
	kid := "#key1"
	payload := encodeDidDocument(didDocumentPath, didID)
	signature := "mAJp4ZHwY5UMA05OEKvoZreRo0XrYe77s3RLyGKArG85IoBULs4cLDBtdpOToCtSZhPvCC2xOUXMGyGXDmmEHg"
	return request(alg, kid, payload, signature, operation)
}

func request(alg, kid, payload, signature string, operation model.OperationType) *model.Request {
	header := &model.Header{
		Alg:       alg,
		Kid:       kid,
		Operation: operation,
	}
	req := &model.Request{
		Header:    header,
		Payload:   payload,
		Signature: signature}
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
