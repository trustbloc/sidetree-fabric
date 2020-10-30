/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package dcashandler

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/trustbloc/sidetree-fabric/pkg/httpserver"
	"github.com/trustbloc/sidetree-fabric/pkg/mocks"
)

//go:generate counterfeiter -o ./mocks/ioreader.gen.go --fake-name IOReader io.Reader

func TestNewUploadHandler(t *testing.T) {
	h := NewUploadHandler(channel1, handlerCfg, nil)
	require.NotNil(t, h)

	require.Equal(t, "/cas", h.Path())
	require.Equal(t, http.MethodPost, h.Method())
}

func TestUpload_Handler(t *testing.T) {
	dcasProvider := &mocks.DCASClientProvider{}

	h := NewUploadHandler(channel1, handlerCfg, dcasProvider)
	require.NotNil(t, h)

	t.Run("DCAS provider error -> Server Error", func(t *testing.T) {
		errExpected := errors.New("injected DCAS provider error")
		dcasProvider.GetDCASClientReturns(nil, errExpected)

		rw := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/cas", bytes.NewReader([]byte{1, 2, 3, 4}))
		h.Handler()(rw, req)

		require.Equal(t, http.StatusInternalServerError, rw.Result().StatusCode)
		require.Equal(t, httpserver.ContentTypeText, rw.Header().Get(httpserver.ContentTypeHeader))
		require.Equal(t, httpserver.StatusServerError, rw.Body.String())
	})

	t.Run("DCAS client error -> Server Error", func(t *testing.T) {
		restoreParams := setParams(hash, maxSize)
		defer restoreParams()

		errExpected := errors.New("injected DCAS client error")
		dcasClient := &mocks.DCASClient{}
		dcasClient.PutReturns("", errExpected)
		dcasProvider.GetDCASClientReturns(dcasClient, nil)

		rw := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/cas", bytes.NewReader([]byte{1, 2, 3, 4}))
		h.Handler()(rw, req)

		require.Equal(t, http.StatusInternalServerError, rw.Result().StatusCode)
		require.Equal(t, httpserver.ContentTypeText, rw.Header().Get(httpserver.ContentTypeHeader))
		require.Equal(t, httpserver.StatusServerError, rw.Body.String())
	})

	t.Run("Success", func(t *testing.T) {
		content := []byte{1, 2, 3, 4}

		restoreParams := setParams(hash, maxSize)
		defer restoreParams()

		dcasClient := &mocks.DCASClient{}
		dcasClient.PutReturns(hash, nil)
		dcasProvider.GetDCASClientReturns(dcasClient, nil)

		rw := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/cas", bytes.NewReader(content))
		h.Handler()(rw, req)

		require.Equal(t, http.StatusOK, rw.Result().StatusCode)
		require.Equal(t, httpserver.ContentTypeJSON, rw.Header().Get(httpserver.ContentTypeHeader))
		require.Equal(t, fmt.Sprintf(`{"hash":"%s"}`, hash), rw.Body.String())
	})
}

func TestUploadWriter_Write(t *testing.T) {
	const hash = "1234"

	t.Run("Success", func(t *testing.T) {
		r := httptest.NewRecorder()
		rw := newUploadWriter(r)
		require.NotNil(t, rw)

		rw.Write(hash)

		require.Equal(t, http.StatusOK, r.Result().StatusCode)
		require.Equal(t, httpserver.ContentTypeJSON, r.Header().Get(httpserver.ContentTypeHeader))
		require.Equal(t, fmt.Sprintf(`{"hash":"%s"}`, hash), r.Body.String())
	})

	t.Run("Marshal error", func(t *testing.T) {
		r := httptest.NewRecorder()
		rw := newUploadWriter(r)
		require.NotNil(t, rw)

		rw.jsonMarshal = func(v interface{}) ([]byte, error) { return nil, errors.New("injected marshal error") }
		rw.Write(hash)

		require.Equal(t, http.StatusInternalServerError, r.Result().StatusCode)
		require.Equal(t, httpserver.ContentTypeText, r.Header().Get(httpserver.ContentTypeHeader))
		require.Equal(t, httpserver.StatusServerError, r.Body.String())
	})
}
