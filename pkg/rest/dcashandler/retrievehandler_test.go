/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package dcashandler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/trustbloc/sidetree-fabric/pkg/httpserver"
	"github.com/trustbloc/sidetree-fabric/pkg/mocks"
)

const (
	channel1 = "channel1"
	cc1      = "cc1"
	coll1    = "coll1"
	hash     = "123456"
	maxSize  = "1024"
)

var (
	handlerCfg = Config{
		BasePath:      "/cas",
		ChaincodeName: cc1,
		Collection:    coll1,
	}
)

func TestNewRetrieveHandler(t *testing.T) {
	h := NewRetrieveHandler(channel1, handlerCfg, nil)
	require.NotNil(t, h)

	require.Equal(t, "/cas/{hash}", h.Path())
	require.Equal(t, http.MethodGet, h.Method())
	require.NotEmpty(t, h.Params()[maxSizeParam])
}

func TestRetrieve_Handler(t *testing.T) {
	dcasProvider := &mocks.DCASClientProvider{}

	h := NewRetrieveHandler(channel1, handlerCfg, dcasProvider)
	require.NotNil(t, h)

	t.Run("DCAS provider error -> Server Error", func(t *testing.T) {
		restoreParams := setParams(hash, maxSize)
		defer restoreParams()

		errExpected := errors.New("injected DCAS provider error")
		dcasProvider.ForChannelReturns(nil, errExpected)

		rw := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/cas", nil)
		h.Handler()(rw, req)

		require.Equal(t, http.StatusInternalServerError, rw.Result().StatusCode)
		require.Equal(t, httpserver.ContentTypeText, rw.Header().Get(httpserver.ContentTypeHeader))
		require.Equal(t, CodeCasNotReachable, rw.Body.String())
	})

	t.Run("DCAS client error -> Server Error", func(t *testing.T) {
		restoreParams := setParams(hash, maxSize)
		defer restoreParams()

		errExpected := errors.New("injected DCAS client error")
		dcasClient := &mocks.DCASClient{}
		dcasClient.GetReturns(nil, errExpected)
		dcasProvider.ForChannelReturns(dcasClient, nil)

		rw := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/cas", nil)
		h.Handler()(rw, req)

		require.Equal(t, http.StatusInternalServerError, rw.Result().StatusCode)
		require.Equal(t, httpserver.ContentTypeText, rw.Header().Get(httpserver.ContentTypeHeader))
		require.Equal(t, CodeCasNotReachable, rw.Body.String())
	})

	t.Run("No hash -> Bad Request", func(t *testing.T) {
		rw := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/cas", nil)
		h.Handler()(rw, req)

		require.Equal(t, http.StatusBadRequest, rw.Result().StatusCode)
		require.Equal(t, httpserver.ContentTypeText, rw.Header().Get(httpserver.ContentTypeHeader))
		require.Equal(t, CodeInvalidHash, rw.Body.String())
	})

	t.Run("No max-size -> Bad Request", func(t *testing.T) {
		restoreParams := setParams(hash, "0")
		defer restoreParams()

		rw := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/cas", nil)
		h.Handler()(rw, req)

		require.Equal(t, http.StatusBadRequest, rw.Result().StatusCode)
		require.Equal(t, httpserver.ContentTypeText, rw.Header().Get(httpserver.ContentTypeHeader))
		require.Equal(t, CodeMaxSizeNotSpecified, rw.Body.String())
	})

	t.Run("Invalid max-size -> Bad Request", func(t *testing.T) {
		restoreParams := setParams(hash, "xxx")
		defer restoreParams()

		rw := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/cas", nil)
		h.Handler()(rw, req)

		require.Equal(t, http.StatusBadRequest, rw.Result().StatusCode)
		require.Equal(t, httpserver.ContentTypeText, rw.Header().Get(httpserver.ContentTypeHeader))
		require.Equal(t, CodeMaxSizeNotSpecified, rw.Body.String())
	})

	t.Run("No content -> Not Found", func(t *testing.T) {
		restoreParams := setParams(hash, maxSize)
		defer restoreParams()

		dcasClient := &mocks.DCASClient{}
		dcasProvider.ForChannelReturns(dcasClient, nil)

		rw := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/cas", nil)
		h.Handler()(rw, req)

		require.Equal(t, http.StatusNotFound, rw.Result().StatusCode)
		require.Equal(t, httpserver.ContentTypeText, rw.Header().Get(httpserver.ContentTypeHeader))
		require.Equal(t, CodeNotFound, rw.Body.String())
	})

	t.Run("Success", func(t *testing.T) {
		content := []byte{1, 2, 3, 4}

		restoreParams := setParams(hash, maxSize)
		defer restoreParams()

		dcasClient := &mocks.DCASClient{}
		dcasClient.GetReturns(content, nil)
		dcasProvider.ForChannelReturns(dcasClient, nil)

		rw := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/cas", nil)
		h.Handler()(rw, req)

		require.Equal(t, http.StatusOK, rw.Result().StatusCode)
		require.Equal(t, httpserver.ContentTypeBinary, rw.Header().Get(httpserver.ContentTypeHeader))
		require.Equal(t, content, rw.Body.Bytes())
	})

	t.Run("max-size exceeded -> error", func(t *testing.T) {
		content := []byte{1, 2, 3, 4}

		restoreParams := setParams(hash, "1")
		defer restoreParams()

		dcasClient := &mocks.DCASClient{}
		dcasClient.GetReturns(content, nil)
		dcasProvider.ForChannelReturns(dcasClient, nil)

		rw := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/cas", nil)
		h.Handler()(rw, req)

		require.Equal(t, http.StatusBadRequest, rw.Result().StatusCode)
		require.Equal(t, httpserver.ContentTypeText, rw.Header().Get(httpserver.ContentTypeHeader))
		require.Equal(t, CodeMaxSizeExceeded, rw.Body.String())
	})
}

func TestRetrieveWriter_Write(t *testing.T) {
	content := []byte{1, 2, 3, 4}

	r := httptest.NewRecorder()
	rw := newRetrieveWriter(r)
	require.NotNil(t, rw)

	rw.Write(content)

	require.Equal(t, http.StatusOK, r.Result().StatusCode)
	require.Equal(t, httpserver.ContentTypeBinary, r.Header().Get(httpserver.ContentTypeHeader))
	require.Equal(t, content, r.Body.Bytes())
}

func TestRetrieveWriter_WriteError(t *testing.T) {
	t.Run("General error", func(t *testing.T) {
		r := httptest.NewRecorder()
		rw := newRetrieveWriter(r)
		require.NotNil(t, rw)

		rw.WriteError(errors.New("injected error"))

		require.Equal(t, http.StatusInternalServerError, r.Result().StatusCode)
		require.Equal(t, httpserver.ContentTypeText, r.Header().Get(httpserver.ContentTypeHeader))
		require.Equal(t, httpserver.StatusServerError, r.Body.String())
	})

	t.Run("Read error", func(t *testing.T) {
		r := httptest.NewRecorder()
		rw := newRetrieveWriter(r)
		require.NotNil(t, rw)

		rw.WriteError(newRetrieveError(http.StatusBadRequest, CodeInvalidHash))

		require.Equal(t, http.StatusBadRequest, r.Result().StatusCode)
		require.Equal(t, httpserver.ContentTypeText, r.Header().Get(httpserver.ContentTypeHeader))
		require.Equal(t, CodeInvalidHash, r.Body.String())
	})
}

func setParams(hash, maxSize string) func() {
	restoreHash := getHash
	restoreMaxSize := getMaxSize

	getHash = func(req *http.Request) string { return hash }
	getMaxSize = func(req *http.Request) int { return maxSizeFromString(maxSize) }

	return func() {
		getHash = restoreHash
		getMaxSize = restoreMaxSize
	}
}
