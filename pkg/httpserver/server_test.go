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

	"github.com/trustbloc/sidetree-core-go/pkg/api/protocol"
	"github.com/trustbloc/sidetree-core-go/pkg/commitment"
	"github.com/trustbloc/sidetree-core-go/pkg/document"
	"github.com/trustbloc/sidetree-core-go/pkg/docutil"
	"github.com/trustbloc/sidetree-core-go/pkg/jws"
	coremocks "github.com/trustbloc/sidetree-core-go/pkg/mocks"
	"github.com/trustbloc/sidetree-core-go/pkg/restapi/common"
	"github.com/trustbloc/sidetree-core-go/pkg/restapi/diddochandler"
	"github.com/trustbloc/sidetree-core-go/pkg/restapi/dochandler"
	"github.com/trustbloc/sidetree-core-go/pkg/versions/1_0/client"
	"github.com/trustbloc/sidetree-core-go/pkg/versions/1_0/model"
	"github.com/trustbloc/sidetree-core-go/pkg/versions/1_0/operationparser"

	"github.com/trustbloc/sidetree-fabric/pkg/mocks"
)

const (
	url       = "localhost:8080"
	clientURL = "http://" + url

	didDocNamespace = "did:sidetree"
	didDocPath      = "/document"

	operationEndpoint  = didDocPath + "/operations"
	resolutionEndpoint = didDocPath + "/identifiers"

	sampleNamespace = "sample:sidetree"
	samplePath      = "/sample"
	sha2_256        = 18
)

var p = protocol.Protocol{
	GenesisTime:                  0,
	MultihashAlgorithms:          []uint{sha2_256},
	MaxOperationCount:            2,
	MaxOperationSize:             1024,
	CompressionAlgorithm:         "GZIP",
	MaxChunkFileSize:             1024,
	MaxProvisionalIndexFileSize:  1024,
	MaxCoreIndexFileSize:         1024,
	SignatureAlgorithms:          []string{"EdDSA", "ES256"},
	KeyAlgorithms:                []string{"Ed25519", "P-256"},
	Patches:                      []string{"add-public-keys", "remove-public-keys", "add-service-endpoints", "remove-service-endpoints", "ietf-json-patch"},
	NonceSize:                    16,
	MaxMemoryDecompressionFactor: 3,
}

func TestServer_Start(t *testing.T) {
	pc := &mocks.ProtocolClient{}
	opp := operationparser.New(p)

	pv := &coremocks.ProtocolVersion{}
	pv.OperationParserReturns(opp)

	pc.CurrentReturns(pv, nil)
	pc.GetReturns(pv, nil)

	didDocHandler := coremocks.NewMockDocumentHandler().WithNamespace(didDocNamespace).WithProtocolClient(pc)
	sampleDocHandler := coremocks.NewMockDocumentHandler().WithNamespace(sampleNamespace).WithProtocolClient(pc)

	s := New(url,
		"",
		"",
		diddochandler.NewUpdateHandler(operationEndpoint, didDocHandler, pc),
		diddochandler.NewResolveHandler(resolutionEndpoint, didDocHandler),
		newSampleUpdateHandler(sampleDocHandler),
		newSampleResolveHandler(sampleDocHandler),
	)
	require.NoError(t, s.Start())
	require.Error(t, s.Start())

	request, err := getCreateRequest()
	require.NoError(t, err)

	var createReq model.CreateRequest
	err = json.Unmarshal(request, &createReq)
	require.NoError(t, err)

	didID, err := docutil.CalculateID(didDocNamespace, createReq.SuffixData, sha2_256)
	require.NoError(t, err)

	sampleID, err := docutil.CalculateID(sampleNamespace, createReq.SuffixData, sha2_256)
	require.NoError(t, err)

	// Wait for the service to start
	time.Sleep(time.Second)

	t.Run("DID doc", func(t *testing.T) {
		resp, err := httpPut(t, clientURL+operationEndpoint, request)
		require.NoError(t, err)
		require.NotEmpty(t, resp)

		var createdDoc document.ResolutionResult
		require.NoError(t, json.Unmarshal(resp, &createdDoc))
		require.Equal(t, didID, createdDoc.Document["id"])

		resp, err = httpGet(t, clientURL+resolutionEndpoint+"/"+didID)
		require.NoError(t, err)
		require.NotEmpty(t, resp)

		var resolvedDoc document.ResolutionResult
		require.NoError(t, json.Unmarshal(resp, &resolvedDoc))
		require.Equal(t, didID, resolvedDoc.Document["id"])
	})
	t.Run("Sample doc", func(t *testing.T) {
		resp, err := httpPut(t, clientURL+samplePath, request)
		require.NoError(t, err)
		require.NotEmpty(t, resp)

		var createdDoc document.ResolutionResult
		require.NoError(t, json.Unmarshal(resp, &createdDoc))
		require.Equal(t, sampleID, createdDoc.Document["id"])

		resp, err = httpGet(t, clientURL+samplePath+"/"+sampleID+"?max-size=1024")
		require.NoError(t, err)
		require.NotEmpty(t, resp)

		var resolvedDoc document.ResolutionResult
		require.NoError(t, json.Unmarshal(resp, &resolvedDoc))
		require.Equal(t, sampleID, resolvedDoc.Document["id"])
	})
	t.Run("Stop", func(t *testing.T) {
		require.NoError(t, s.Stop(context.Background()))
		require.Error(t, s.Stop(context.Background()))
	})
}

func TestServer_RetryOnStartup(t *testing.T) {
	pc := &mocks.ProtocolClient{}
	opp := operationparser.New(p)

	pv := &coremocks.ProtocolVersion{}
	pv.OperationParserReturns(opp)

	pc.CurrentReturns(pv, nil)
	pc.GetReturns(pv, nil)

	didDocHandler := coremocks.NewMockDocumentHandler().WithNamespace(didDocNamespace)
	sampleDocHandler := coremocks.NewMockDocumentHandler().WithNamespace(sampleNamespace)

	s1 := New(url,
		"",
		"",
		diddochandler.NewUpdateHandler(operationEndpoint, didDocHandler, pc),
		diddochandler.NewResolveHandler(resolutionEndpoint, didDocHandler),
		newSampleUpdateHandler(sampleDocHandler),
		newSampleResolveHandler(sampleDocHandler),
	)

	s2 := New(url,
		"",
		"",
		diddochandler.NewUpdateHandler(operationEndpoint, didDocHandler, pc),
		diddochandler.NewResolveHandler(resolutionEndpoint, didDocHandler),
		newSampleUpdateHandler(sampleDocHandler),
		newSampleResolveHandler(sampleDocHandler),
	)

	s3 := New(url,
		"",
		"",
		diddochandler.NewUpdateHandler(operationEndpoint, didDocHandler, pc),
		diddochandler.NewResolveHandler(resolutionEndpoint, didDocHandler),
		newSampleUpdateHandler(sampleDocHandler),
		newSampleResolveHandler(sampleDocHandler),
	)

	// Start three HTTP servers (all listening on the same port) to test the retry logic
	require.NoError(t, s1.Start())
	require.NoError(t, s2.Start())
	require.NoError(t, s3.Start())
	time.Sleep(500 * time.Millisecond)

	require.NoError(t, s1.Stop(context.Background()))
	require.NoError(t, s2.Stop(context.Background()))
	require.NoError(t, s3.Stop(context.Background()))
	time.Sleep(500 * time.Millisecond)

	// Wait for the service to start
	time.Sleep(time.Second)
}

// httpPut sends a regular POST request to the sidetree-node
// - If post request has operation "create" then return sidetree document else no response
func httpPut(t *testing.T, url string, request []byte) ([]byte, error) {
	client := &http.Client{}

	httpReq, err := http.NewRequest("POST", url, bytes.NewReader(request))
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
	pc := &mocks.ProtocolClient{}
	opp := operationparser.New(p)

	pv := &coremocks.ProtocolVersion{}
	pv.OperationParserReturns(opp)

	pc.CurrentReturns(pv, nil)
	pc.GetReturns(pv, nil)

	return &sampleUpdateHandler{
		UpdateHandler: dochandler.NewUpdateHandler(processor, pc),
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

// Params returns the context path
func (h *sampleResolveHandler) Params() map[string]string {
	return map[string]string{"max-size": "{max-size:[0-9]+}"}
}

// Method returns the HTTP method
func (h *sampleResolveHandler) Method() string {
	return http.MethodGet
}

// Handler returns the handler
func (h *sampleResolveHandler) Handler() common.HTTPRequestHandler {
	return h.Resolve
}

func getCreateRequest() ([]byte, error) {
	recoveryKey := &jws.JWK{
		Crv: "crv",
		Kty: "kty",
		X:   "x",
		Y:   "y",
	}

	updateKey := &jws.JWK{
		Crv: "crv",
		Kty: "kty",
		X:   "x",
	}

	recoveryCommitment, err := commitment.GetCommitment(recoveryKey, sha2_256)
	if err != nil {
		return nil, err
	}

	updateCommitment, err := commitment.GetCommitment(updateKey, sha2_256)
	if err != nil {
		return nil, err
	}

	info := &client.CreateRequestInfo{
		OpaqueDocument:     validDoc,
		RecoveryCommitment: recoveryCommitment,
		UpdateCommitment:   updateCommitment,
		MultihashCode:      sha2_256,
	}

	return client.NewCreateRequest(info)
}

const validDoc = `{
	"publicKey": [{
      	"id": "dual-key",
      	"type": "JwsVerificationKey2020",
      	"purposes": ["authentication"],
      	"publicKeyJwk": {
        	"kty": "EC",
        	"crv": "P-256K",
        	"x": "PUymIqdtF_qxaAqPABSw-C-owT1KYYQbsMKFM-L9fJA",
        	"y": "nM84jDHCMOTGTh_ZdHq4dBBdo4Z5PkEOW9jA8z8IsGc"
      	}
	}]
}`
