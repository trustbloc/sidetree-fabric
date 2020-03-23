/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package bddtests

import (
	"net/http"
	"time"

	"github.com/cucumber/godog"
	"github.com/pkg/errors"
	"github.com/trustbloc/fabric-peer-test-common/bddtests"
	"github.com/trustbloc/sidetree-fabric/test/bddtests/restclient"
)

// FileHandlerSteps
type FileHandlerSteps struct {
	resp *restclient.HttpResponse
}

// NewFileHandlerSteps
func NewFileHandlerSteps(_ *bddtests.BDDContext) *FileHandlerSteps {
	return &FileHandlerSteps{}
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

func (d *FileHandlerSteps) checkErrorResponse(statusCode int, msg string) error {
	if d.resp.StatusCode != statusCode {
		return errors.Errorf("expecting status code %d but got %d", statusCode, d.resp.StatusCode)
	}

	if d.resp.ErrorMsg != msg {
		return errors.Errorf("expecting error message [%s] but got [%s]", msg, d.resp.ErrorMsg)
	}

	return nil
}

// RegisterSteps registers did sidetree steps
func (d *FileHandlerSteps) RegisterSteps(s *godog.Suite) {
	s.Step(`^client sends request to "([^"]*)" to retrieve file$`, d.resolveFile)
	s.Step(`^the response has status code (\d+) and error message "([^"]*)"$`, d.checkErrorResponse)
}
