/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package bddtests

import (
	"encoding/base64"
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
	"github.com/trustbloc/sidetree-fabric/test/bddtests/restclient"
)

var logger = logrus.New()

const (
	sha2_256           = 18
	initialValuesParam = ";initial-values="

	updateOTP   = "updateOTP"
	recoveryOTP = "recoveryOTP"
)

// DIDSideSteps
type DIDSideSteps struct {
	encodedCreatePayload string
	reqNamespace         string
	resp                 *restclient.HttpResponse
	bddContext           *bddtests.BDDContext
}

// NewDIDSideSteps
func NewDIDSideSteps(context *bddtests.BDDContext) *DIDSideSteps {
	return &DIDSideSteps{bddContext: context}
}

func (d *DIDSideSteps) sendDIDDocument(url, didDocumentPath, namespace string) error {
	logger.Infof("Creating DID document at %s", url)

	opaqueDoc := getOpaqueDocument(didDocumentPath)
	payload, err := getCreatePayload(opaqueDoc)
	if err != nil {
		return err
	}

	d.encodedCreatePayload = docutil.EncodeToString(payload)
	req, err := getRequest(d.encodedCreatePayload)
	if err != nil {
		return err
	}

	d.reqNamespace = namespace

	d.resp, err = restclient.SendRequest(url, req)
	return err
}

func (d *DIDSideSteps) resolveDIDDocumentWithInitialValue(url string) error {
	did, err := docutil.CalculateID(d.reqNamespace, d.encodedCreatePayload, sha2_256)
	if err != nil {
		return err
	}

	req := url + "/" + did + initialValuesParam + d.encodedCreatePayload
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
	documentHash, err := docutil.CalculateID(d.reqNamespace, d.encodedCreatePayload, sha2_256)
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

func (d *DIDSideSteps) resolveDIDDocument(url string) error {
	documentHash, err := docutil.CalculateID(d.reqNamespace, d.encodedCreatePayload, sha2_256)
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

func (d *DIDSideSteps) deleteDIDDocument(url string) error {
	uniqueSuffix, err := docutil.CalculateUniqueSuffix(d.encodedCreatePayload, sha2_256)
	if err != nil {
		return err
	}

	logger.Infof("delete did document [%s]from %s", uniqueSuffix, url)

	payload, err := getDeletePayload(uniqueSuffix)
	if err != nil {
		return err
	}

	req, err := getRequest(payload)
	if err != nil {
		return err
	}

	d.resp, err = restclient.SendRequest(url, req)

	logger.Infof("delete status %d, error '%s:'", d.resp.StatusCode, d.resp.ErrorMsg)

	return err
}

func getCreatePayload(doc string) ([]byte, error) {
	return helper.NewCreateRequest(&helper.CreateRequestInfo{
		OpaqueDocument:  doc,
		RecoveryKey:     "recoveryKey",
		NextRecoveryOTP: recoveryOTP,
		MultihashCode:   sha2_256,
	})
}

func getRequest(payload string) ([]byte, error) {
	return helper.NewSignedRequest(&helper.SignedRequestInfo{
		Payload:   payload,
		Algorithm: "alg",
		KID:       "kid",
		Signature: "signature",
	})
}

func getDeletePayload(did string) (string, error) {
	return helper.NewDeletePayload(&helper.DeletePayloadInfo{
		DidUniqueSuffix: did,
		RecoveryOTP:     base64.URLEncoding.EncodeToString([]byte(recoveryOTP)),
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
	s.Step(`^client sends request to "([^"]*)" to delete DID document$`, d.deleteDIDDocument)
	s.Step(`^client sends request to "([^"]*)" to resolve DID document with initial value$`, d.resolveDIDDocumentWithInitialValue)
	s.Step(`^check success response contains "([^"]*)"$`, d.checkSuccessResp)
	s.Step(`^client sends request to "([^"]*)" to resolve DID document$`, d.resolveDIDDocument)
}
