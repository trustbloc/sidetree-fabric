/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package blockchainhandler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/trustbloc/sidetree-fabric/pkg/httpserver"
)

func TestResponseWriter(t *testing.T) {
	t.Run("Write", func(t *testing.T) {
		rw := httptest.NewRecorder()
		bcrw := newBlockchainWriter(rw)
		require.NotNil(t, bcrw)

		content := []byte(`{"field":"value"}`)
		bcrw.Write(content)

		require.Equal(t, http.StatusOK, rw.Result().StatusCode)
		require.Equal(t, httpserver.ContentTypeJSON, rw.Header().Get(httpserver.ContentTypeHeader))
		require.Equal(t, content, rw.Body.Bytes())
	})

	t.Run("Write error", func(t *testing.T) {
		rw := httptest.NewRecorder()
		bcrw := newBlockchainWriter(rw)
		require.NotNil(t, bcrw)

		bcrw.WriteError(errors.New("some error"))

		require.Equal(t, http.StatusInternalServerError, rw.Result().StatusCode)
		require.Equal(t, httpserver.ContentTypeText, rw.Header().Get(httpserver.ContentTypeHeader))
		require.Equal(t, httpserver.StatusServerError, rw.Body.String())
	})

	t.Run("Write blockchain error", func(t *testing.T) {
		rw := httptest.NewRecorder()
		bcrw := newBlockchainWriter(rw)
		require.NotNil(t, bcrw)

		bcrw.WriteError(newBadRequestError(InvalidTxNumOrTimeHash))

		require.Equal(t, http.StatusBadRequest, rw.Result().StatusCode)
		require.Equal(t, httpserver.ContentTypeJSON, rw.Header().Get(httpserver.ContentTypeHeader))

		errResp := &ErrorResponse{}
		require.NoError(t, json.Unmarshal(rw.Body.Bytes(), errResp))
		require.Equal(t, InvalidTxNumOrTimeHash, errResp.Code)
	})

	t.Run("Marshal error", func(t *testing.T) {
		rw := httptest.NewRecorder()
		bcrw := newBlockchainWriter(rw)
		require.NotNil(t, bcrw)

		bcrw.jsonMarshal = func(v interface{}) ([]byte, error) { return nil, errors.New("injected marshal error") }

		err := newBadRequestError(InvalidTxNumOrTimeHash)
		bcrw.WriteError(err)

		require.Equal(t, http.StatusBadRequest, rw.Result().StatusCode)
		require.Equal(t, httpserver.ContentTypeText, rw.Header().Get(httpserver.ContentTypeHeader))
		require.Equal(t, err.Error(), rw.Body.String())
	})
}
