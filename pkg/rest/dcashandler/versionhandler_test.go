/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package dcashandler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/trustbloc/sidetree-fabric/pkg/httpserver"
)

func TestNewVersionHandler(t *testing.T) {
	h := NewVersionHandler(channel1, handlerCfg)
	require.NotNil(t, h)

	require.Equal(t, "/cas/version", h.Path())
	require.Equal(t, http.MethodGet, h.Method())
}

func TestVersion_Handler(t *testing.T) {
	const v1 = "1.0.1"

	cfg := Config{
		BasePath:      "/cas",
		ChaincodeName: cc1,
		Collection:    coll1,
		Version:       v1,
	}

	t.Run("Success", func(t *testing.T) {
		h := NewVersionHandler(channel1, cfg)
		require.NotNil(t, h)

		rw := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/cas/version", nil)

		h.Handler()(rw, req)

		require.Equal(t, http.StatusOK, rw.Result().StatusCode)
		require.Equal(t, httpserver.ContentTypeJSON, rw.Header().Get(httpserver.ContentTypeHeader))

		resp := &VersionResponse{}
		require.NoError(t, json.Unmarshal(rw.Body.Bytes(), resp))
		require.Equal(t, moduleName, resp.Name)
		require.Equal(t, v1, resp.Version)
	})

	t.Run("Marshal error -> Server Error", func(t *testing.T) {
		h := NewVersionHandler(channel1, cfg)
		require.NotNil(t, h)
		h.jsonMarshal = func(v interface{}) ([]byte, error) { return nil, errors.New("injected marshal error") }

		rw := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/cas/version", nil)

		h.Handler()(rw, req)

		require.Equal(t, http.StatusInternalServerError, rw.Result().StatusCode)
		require.Equal(t, httpserver.ContentTypeText, rw.Header().Get(httpserver.ContentTypeHeader))
		require.Equal(t, httpserver.StatusServerError, rw.Body.String())
	})
}
