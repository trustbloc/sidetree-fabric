/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package httpserver

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/trustbloc/sidetree-core-go/pkg/document"
	"github.com/trustbloc/sidetree-core-go/pkg/mocks"
	"github.com/trustbloc/sidetree-core-go/pkg/restapi/common"
	"github.com/trustbloc/sidetree-core-go/pkg/restapi/diddochandler"
	"github.com/trustbloc/sidetree-core-go/pkg/restapi/dochandler"
	"github.com/trustbloc/sidetree-core-go/pkg/restapi/model"
)

const (
	url       = "localhost:8080"
	clientURL = "http://" + url

	didDocNamespace = "did:sidetree:"
	didID           = didDocNamespace + "EiDOQXC2GnoVyHwIRbjhLx_cNc6vmZaS04SZjZdlLLAPRg=="

	sampleNamespace = "sample:sidetree:"
	samplePath      = "/sample"
	sampleID        = sampleNamespace + "EiDOQXC2GnoVyHwIRbjhLx_cNc6vmZaS04SZjZdlLLAPRg=="

	createRequest = `{
  "header": {
    "operation": "create",
    "kid": "#key1",
    "alg": "ES256K"
  },
  "payload": "ewogICJAY29udGV4dCI6ICJodHRwczovL3czaWQub3JnL2RpZC92MSIsCiAgInB1YmxpY0tleSI6IFt7CiAgICAiaWQiOiAiI2tleTEiLAogICAgInR5cGUiOiAiU2VjcDI1NmsxVmVyaWZpY2F0aW9uS2V5MjAxOCIsCiAgICAicHVibGljS2V5SGV4IjogIjAyZjQ5ODAyZmIzZTA5YzZkZDQzZjE5YWE0MTI5M2QxZTBkYWQwNDRiNjhjZjgxY2Y3MDc5NDk5ZWRmZDBhYTlmMSIKICB9XSwKICAic2VydmljZSI6IFt7CiAgICAiaWQiOiAiSWRlbnRpdHlIdWIiLAogICAgInR5cGUiOiAiSWRlbnRpdHlIdWIiLAogICAgInNlcnZpY2VFbmRwb2ludCI6IHsKICAgICAgIkBjb250ZXh0IjogInNjaGVtYS5pZGVudGl0eS5mb3VuZGF0aW9uL2h1YiIsCiAgICAgICJAdHlwZSI6ICJVc2VyU2VydmljZUVuZHBvaW50IiwKICAgICAgImluc3RhbmNlIjogWyJkaWQ6YmFyOjQ1NiIsICJkaWQ6emF6Ojc4OSJdCiAgICB9CiAgfV0KfQo=",
  "signature": "mAJp4ZHwY5UMA05OEKvoZreRo0XrYe77s3RLyGKArG85IoBULs4cLDBtdpOToCtSZhPvCC2xOUXMGyGXDmmEHg"
}
`
)

func TestServer_Start(t *testing.T) {
	didDocHandler := mocks.NewMockDocumentHandler().WithNamespace(didDocNamespace)
	sampleDocHandler := mocks.NewMockDocumentHandler().WithNamespace(sampleNamespace)

	s := New(url,
		diddochandler.NewUpdateHandler(didDocHandler),
		diddochandler.NewResolveHandler(didDocHandler),
		newSampleUpdateHandler(sampleDocHandler),
		newSampleResolveHandler(sampleDocHandler),
	)
	require.NoError(t, s.Start())
	require.Error(t, s.Start())

	// Wait for the service to start
	time.Sleep(time.Second)

	t.Run("DID doc", func(t *testing.T) {
		request := &model.Request{}
		err := json.Unmarshal([]byte(createRequest), request)
		require.NoError(t, err)

		resp, err := httpPut(t, clientURL+diddochandler.Path, request)
		require.NoError(t, err)
		require.NotEmpty(t, resp)

		var createdDoc document.Document
		require.NoError(t, json.Unmarshal(resp, &createdDoc))
		require.Equal(t, didID, createdDoc["id"])

		resp, err = httpGet(t, clientURL+diddochandler.Path+"/"+didID)
		require.NoError(t, err)
		require.NotEmpty(t, resp)

		var resolvedDoc document.Document
		require.NoError(t, json.Unmarshal(resp, &resolvedDoc))
		require.Equal(t, didID, resolvedDoc["id"])
	})
	t.Run("Sample doc", func(t *testing.T) {
		request := &model.Request{}
		err := json.Unmarshal([]byte(createRequest), request)
		require.NoError(t, err)

		resp, err := httpPut(t, clientURL+samplePath, request)
		require.NoError(t, err)
		require.NotEmpty(t, resp)

		var createdDoc document.Document
		require.NoError(t, json.Unmarshal(resp, &createdDoc))
		require.Equal(t, sampleID, createdDoc["id"])

		resp, err = httpGet(t, clientURL+samplePath+"/"+sampleID)
		require.NoError(t, err)
		require.NotEmpty(t, resp)

		var resolvedDoc document.Document
		require.NoError(t, json.Unmarshal(resp, &resolvedDoc))
		require.Equal(t, sampleID, resolvedDoc["id"])
	})
	t.Run("Stop", func(t *testing.T) {
		require.NoError(t, s.Stop(context.Background()))
		require.Error(t, s.Stop(context.Background()))
	})
}

// httpPut sends a regular POST request to the sidetree-node
// - If post request has operation "create" then return sidetree document else no response
func httpPut(t *testing.T, url string, req *model.Request) ([]byte, error) {
	client := &http.Client{}
	b, err := json.Marshal(req)
	require.NoError(t, err)

	httpReq, err := http.NewRequest("POST", url, bytes.NewReader(b))
	require.NoError(t, err)

	httpReq.Header.Set("Content-Type", "application/did+ld+json")
	resp, err := invokeWithRetry(
		func() (response *http.Response, e error) {
			return client.Do(httpReq)
		},
	)
	require.NoError(t, err)
	require.Equal(t, "application/did+ld+json", resp.Header.Get("content-type"))
	return handleHttpResp(t, resp)
}

// httpGet send a regular GET request to the sidetree-node and expects 'side tree document' argument as a response
func httpGet(t *testing.T, url string) ([]byte, error) {
	client := &http.Client{}
	resp, err := invokeWithRetry(
		func() (response *http.Response, e error) {
			return client.Get(url)
		},
	)
	require.NoError(t, err)
	return handleHttpResp(t, resp)
}

func handleHttpResp(t *testing.T, resp *http.Response) ([]byte, error) {
	if status := resp.StatusCode; status != http.StatusOK {
		return nil, fmt.Errorf(string(read(t, resp)))
	}
	return read(t, resp), nil
}

func read(t *testing.T, response *http.Response) []byte {
	respBytes, err := ioutil.ReadAll(response.Body)
	require.NoError(t, err)
	return respBytes
}

func invokeWithRetry(invoke func() (*http.Response, error)) (*http.Response, error) {
	remainingAttempts := 20
	for {
		resp, err := invoke()
		if err == nil {
			return resp, err
		}
		remainingAttempts--
		if remainingAttempts == 0 {
			return nil, err
		}
		time.Sleep(100 * time.Millisecond)
	}
}

type sampleUpdateHandler struct {
	*dochandler.UpdateHandler
}

func newSampleUpdateHandler(processor dochandler.Processor) *sampleUpdateHandler {
	return &sampleUpdateHandler{
		UpdateHandler: dochandler.NewUpdateHandler(processor),
	}
}

// Path returns the context path
func (h *sampleUpdateHandler) Path() string {
	return samplePath
}

// Method returns the HTTP method
func (h *sampleUpdateHandler) Method() string {
	return http.MethodPost
}

// Handler returns the handler
func (h *sampleUpdateHandler) Handler() common.HTTPRequestHandler {
	return h.Update
}

// Update creates/updates the document
func (o *sampleUpdateHandler) Update(rw http.ResponseWriter, req *http.Request) {
	o.UpdateHandler.Update(rw, req)
}

type sampleResolveHandler struct {
	*dochandler.ResolveHandler
}

func newSampleResolveHandler(resolver dochandler.Resolver) *sampleResolveHandler {
	return &sampleResolveHandler{
		ResolveHandler: dochandler.NewResolveHandler(resolver),
	}
}

// Path returns the context path
func (h *sampleResolveHandler) Path() string {
	return samplePath + "/{id}"
}

// Method returns the HTTP method
func (h *sampleResolveHandler) Method() string {
	return http.MethodGet
}

// Handler returns the handler
func (h *sampleResolveHandler) Handler() common.HTTPRequestHandler {
	return h.Resolve
}
