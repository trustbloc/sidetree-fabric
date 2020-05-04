/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package sidetreehandler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/trustbloc/sidetree-fabric/pkg/httpserver"
	"github.com/trustbloc/sidetree-fabric/pkg/rest/versionhandler"
)

const (
	channel1 = "channel1"
	version  = "0.1.3"
)

func TestNewVersionHandler(t *testing.T) {
	handlerCfg := Config{
		BasePath: "/sidetree",
	}

	h := NewVersionHandler(channel1, handlerCfg)
	require.NotNil(t, h)

	require.Equal(t, "/sidetree/version", h.Path())
	require.Equal(t, http.MethodGet, h.Method())
}

func TestVersion_Handler(t *testing.T) {
	handlerCfg := Config{
		BasePath: "/sidetree",
		Version:  version,
	}

	h := NewVersionHandler(channel1, handlerCfg)
	require.NotNil(t, h)

	rw := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/sidetree/version", nil)

	h.Handler()(rw, req)

	require.Equal(t, http.StatusOK, rw.Result().StatusCode)
	require.Equal(t, httpserver.ContentTypeJSON, rw.Header().Get(httpserver.ContentTypeHeader))

	resp := &versionhandler.Response{}
	require.NoError(t, json.Unmarshal(rw.Body.Bytes(), resp))
	require.Equal(t, moduleName, resp.Name)
	require.Equal(t, version, resp.Version)
}
