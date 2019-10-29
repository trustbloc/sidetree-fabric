/*
Copyright SecureKey Technologies Inc. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package restclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/trustbloc/sidetree-core-go/pkg/restapi/model"
)

type HttpResponse struct {
	StatusCode int
	Payload    []byte
	ErrorMsg   string
}

// SendRequest sends a regular POST request to the sidetree-node
// - If post request has operation "create" then return sidetree document else no response
func SendRequest(url string, req *model.Request) (*HttpResponse, error) {
	resp, err := sendHTTPRequest(url, req)
	if err != nil {
		return nil, err
	}
	return handleHttpResp(resp)
}

// SendResolveRequest send a regular GET request to the sidetree-node and expects 'side tree document' argument as a response
func SendResolveRequest(url string) (*HttpResponse, error) {
	client := &http.Client{}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	return handleHttpResp(resp)
}

func handleHttpResp(resp *http.Response) (*HttpResponse, error) {
	gotBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body failed: %w", err)
	}
	if status := resp.StatusCode; status != http.StatusOK {
		return &HttpResponse{
			StatusCode: status,
			ErrorMsg:   string(gotBody),
		}, nil
	}
	return &HttpResponse{StatusCode: http.StatusOK, Payload: gotBody}, nil
}

func sendHTTPRequest(url string, req *model.Request) (*http.Response, error) {
	client := &http.Client{}
	b, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")

	return client.Do(httpReq)
}
