/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package blockchainhandler

import (
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
}
