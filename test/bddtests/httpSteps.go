/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package bddtests

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/trustbloc/fabric-peer-test-common/bddtests"
)

const (
	contentTypeJSON = "application/json"
)

type httpSteps struct {
	response    []byte
	statusCode  int
	contentType string
}

func (d *httpSteps) httpPost(url string, req []byte, contentType string) error {
	logger.Infof("Invoking POST on [%s]", url)

	resp, statusCode, header, err := bddtests.HTTPPost(url, req, contentType)
	if err != nil {
		return err
	}

	d.setResponse(statusCode, resp, header)

	return nil
}

func (d *httpSteps) httpGet(url string) error {
	logger.Infof("Invoking GET on [%s]", url)

	resp, statusCode, header, err := bddtests.HTTPGet(url)
	if err != nil {
		return err
	}

	d.setResponse(statusCode, resp, header)

	return nil
}

func (d *httpSteps) httpGetWithRetry(url string, attempts uint8, retryableCode int, retryableCodes ...int) error {
	logger.Infof("Resolving: %s", url)

	codes := append(retryableCodes, retryableCode)

	remainingAttempts := attempts
	for {
		err := d.httpGet(url)
		if err != nil {
			return err
		}

		if !d.containsStatusCode(codes, d.statusCode) {
			break
		}

		logger.Infof("Not found: %s. Remaining attempts: %d", url, remainingAttempts)

		remainingAttempts--
		if remainingAttempts == 0 {
			break
		}

		time.Sleep(time.Second)
	}

	return nil
}

func (d *httpSteps) containsStatusCode(codes []int, code int) bool {
	for _, c := range codes {
		if c == code {
			return true
		}
	}

	return false
}

func (d *httpSteps) setResponse(statusCode int, response []byte, header http.Header) {
	d.statusCode = statusCode
	d.response = response

	logger.Infof("Got header: %s", header)

	contentType, ok := header["Content-Type"]
	if ok {
		d.contentType = contentType[0]
	}
}

func (d *httpSteps) checkResponse(statusCode int, msg string) error {
	if d.statusCode != statusCode {
		return errors.Errorf("expecting status code %d but got %d", statusCode, d.statusCode)
	}

	respMsg := d.getErrorResponse()
	if respMsg != msg {
		return errors.Errorf("expecting error message [%s] but got [%s]", msg, respMsg)
	}

	return nil
}

func (d *httpSteps) checkErrorResponse(msg string) error {
	if d.statusCode == http.StatusOK {
		return errors.New("expecting error but got OK")
	}

	respMsg := d.getErrorResponse()
	if !strings.Contains(respMsg, msg) {
		return errors.Errorf("expecting error message contains [%s] but got [%s]", msg, respMsg)
	}

	return nil
}

func (d *httpSteps) getErrorResponse() string {
	respMsg := ""

	if d.contentType == contentTypeJSON || d.contentType == "application/did+ld+json" {
		if err := json.Unmarshal(d.response, &respMsg); err != nil {
			respMsg = string(d.response)
		}
	} else {
		respMsg = string(d.response)
	}

	return respMsg
}
