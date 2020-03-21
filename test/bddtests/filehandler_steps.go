/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package bddtests

import (
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/cucumber/godog"
	jsonpatch "github.com/evanphx/json-patch"
	"github.com/pkg/errors"
	"github.com/trustbloc/fabric-peer-test-common/bddtests"
	"github.com/trustbloc/sidetree-core-go/pkg/document"
	"github.com/trustbloc/sidetree-core-go/pkg/docutil"
	"github.com/trustbloc/sidetree-core-go/pkg/restapi/helper"

	"github.com/trustbloc/sidetree-fabric/test/bddtests/restclient"
)

// FileHandlerSteps
type FileHandlerSteps struct {
	encodedCreatePayload string
	reqNamespace         string
	resp                 *restclient.HttpResponse
	bddContext           *bddtests.BDDContext
}

// NewFileHandlerSteps
func NewFileHandlerSteps(context *bddtests.BDDContext) *FileHandlerSteps {
	return &FileHandlerSteps{bddContext: context}
}

func (d *FileHandlerSteps) resolveFile(url string) error {
	logger.Infof("Resolving file: %s", url)

	remainingAttempts := 20
	for {
		var err error
		d.resp, err = restclient.SendResolveRequest(url)
		if err != nil {
			return err
		}
		if d.resp.StatusCode == http.StatusNotFound {
			logger.Infof("File not found: %s. Remaining attempts: %d", url, remainingAttempts)
			remainingAttempts--
			if remainingAttempts > 0 {
				time.Sleep(time.Second)
				continue
			}
		}

		bddtests.SetResponse(string(d.resp.Payload))

		return nil
	}
}

func (d *FileHandlerSteps) checkErrorResp(errorMsg string) error {
	if !strings.Contains(d.resp.ErrorMsg, errorMsg) {
		return errors.Errorf("error resp %s doesn't contain %s", d.resp.ErrorMsg, errorMsg)
	}
	return nil
}

func (d *FileHandlerSteps) checkErrorResponse(statusCode int, msg string) error {
	if d.resp.StatusCode != statusCode {
		return errors.Errorf("expecting status code %d but got %d", statusCode, d.resp.StatusCode)
	}

	if d.resp.ErrorMsg != msg {
		return errors.Errorf("expecting error message [%s] but got [%s]", msg, d.resp.ErrorMsg)
	}

	return nil
}

func (d *FileHandlerSteps) getOpaqueDocument(content string) string {
	doc, _ := document.FromBytes([]byte(content))
	bytes, _ := doc.Bytes()
	return string(bytes)
}

func (d *FileHandlerSteps) getUpdateRequest(uniqueSuffix string, patch jsonpatch.Patch) ([]byte, error) {
	return helper.NewUpdateRequest(&helper.UpdateRequestInfo{
		DidUniqueSuffix: uniqueSuffix,
		Patch:           patch,
		UpdateOTP:       docutil.EncodeToString([]byte(updateOTP)),
		MultihashCode:   sha2_256,
	})
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
	s.Step(`^client sends request to "([^"]*)" to retrieve file$`, d.resolveFile)
	s.Step(`^the response has status code (\d+) and error message "([^"]*)"$`, d.checkErrorResponse)
}
