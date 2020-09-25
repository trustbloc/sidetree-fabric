/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package sidetreehandler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/trustbloc/sidetree-core-go/pkg/api/protocol"
	coremocks "github.com/trustbloc/sidetree-core-go/pkg/mocks"

	"github.com/trustbloc/sidetree-fabric/pkg/httpserver"
	"github.com/trustbloc/sidetree-fabric/pkg/mocks"
	"github.com/trustbloc/sidetree-fabric/pkg/rest/authhandler"
)

const (
	channel1 = "channel1"

	v1 = "0.1"
	v2 = "0.1.1"
)

var handlerCfg = Config{
	Authorization: authhandler.Config{},
	Version:       "0.1.5",
	BasePath:      "/sidetree",
	Namespace:     "did:sidetree",
}

func TestNewHandler(t *testing.T) {
	h := NewVersionHandler(channel1, handlerCfg, &mocks.ProtocolClient{})
	require.NotNil(t, h)

	require.Equal(t, "/sidetree/version", h.Path())
	require.Equal(t, http.MethodGet, h.Method())
}

func Test_Handler(t *testing.T) {
	p1 := protocol.Protocol{
		GenesisTime:       1,
		MaxOperationCount: 10,
		MaxOperationSize:  1000,
	}

	p2 := protocol.Protocol{
		GenesisTime:       100,
		MaxOperationCount: 20,
		MaxOperationSize:  2000,
	}

	pv1 := &coremocks.ProtocolVersion{}
	pv1.VersionReturns(v1)
	pv1.ProtocolReturns(p1)

	pv2 := &coremocks.ProtocolVersion{}
	pv2.VersionReturns(v2)
	pv2.ProtocolReturns(p2)

	pc := &mocks.ProtocolClient{}
	pc.CurrentReturns(pv1, nil)
	pc.GetReturns(pv2, nil)

	h := NewVersionHandler(channel1, handlerCfg, pc)
	require.NotNil(t, h)

	t.Run("Current version -> Success", func(t *testing.T) {
		rw := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, handlerCfg.BasePath, nil)

		h.Handler()(rw, req)

		require.Equal(t, http.StatusOK, rw.Result().StatusCode)
		require.Equal(t, httpserver.ContentTypeJSON, rw.Header().Get(httpserver.ContentTypeHeader))

		var resp response
		require.NoError(t, json.Unmarshal(rw.Body.Bytes(), &resp))
		require.Equal(t, v1, resp.Version)
		require.Equal(t, p1.GenesisTime, resp.GenesisTime)
	})

	t.Run("Protocol at time -> Success", func(t *testing.T) {
		rw := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, handlerCfg.BasePath, nil)

		restore := setTimeParam("200")
		defer restore()

		h.Handler()(rw, req)

		require.Equal(t, http.StatusOK, rw.Result().StatusCode)
		require.Equal(t, httpserver.ContentTypeJSON, rw.Header().Get(httpserver.ContentTypeHeader))

		var resp response
		require.NoError(t, json.Unmarshal(rw.Body.Bytes(), &resp))
		require.Equal(t, v2, resp.Version)
		require.Equal(t, p2.GenesisTime, resp.GenesisTime)
	})
}

func Test_HandlerError(t *testing.T) {
	t.Run("Invalid time error", func(t *testing.T) {
		pc := &mocks.ProtocolClient{}

		h := NewVersionHandler(channel1, handlerCfg, pc)
		require.NotNil(t, h)

		rw := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, handlerCfg.BasePath, nil)

		restore := setTimeParam("-200")
		defer restore()

		h.Handler()(rw, req)

		require.Equal(t, http.StatusBadRequest, rw.Result().StatusCode)
		require.Equal(t, httpserver.ContentTypeText, rw.Header().Get(httpserver.ContentTypeHeader))
		require.Equal(t, "invalid time: -200", rw.Body.String())
	})

	t.Run("Protocol client error", func(t *testing.T) {
		errExpected := fmt.Errorf("injected version client error")

		pc := &mocks.ProtocolClient{}

		h := NewVersionHandler(channel1, handlerCfg, pc)
		require.NotNil(t, h)

		t.Run("Current version", func(t *testing.T) {
			pc.CurrentReturns(nil, errExpected)

			rw := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, handlerCfg.BasePath, nil)

			h.Handler()(rw, req)

			require.Equal(t, http.StatusInternalServerError, rw.Result().StatusCode)
			require.Equal(t, httpserver.ContentTypeText, rw.Header().Get(httpserver.ContentTypeHeader))
			require.Equal(t, httpserver.StatusServerError, rw.Body.String())
		})

		t.Run("Protocol at time not found", func(t *testing.T) {
			pc.GetReturns(nil, fmt.Errorf("version parameters are not defined for blockchain time"))

			rw := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, handlerCfg.BasePath, nil)

			restore := setTimeParam("200")
			defer restore()

			h.Handler()(rw, req)

			require.Equal(t, http.StatusNotFound, rw.Result().StatusCode)
			require.Equal(t, httpserver.ContentTypeText, rw.Header().Get(httpserver.ContentTypeHeader))
			require.Equal(t, "version not found for time: 200", rw.Body.String())
		})

		t.Run("Protocol at time error", func(t *testing.T) {
			pc.GetReturns(nil, errExpected)

			rw := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, handlerCfg.BasePath, nil)

			restore := setTimeParam("200")
			defer restore()

			h.Handler()(rw, req)

			require.Equal(t, http.StatusInternalServerError, rw.Result().StatusCode)
			require.Equal(t, httpserver.ContentTypeText, rw.Header().Get(httpserver.ContentTypeHeader))
			require.Equal(t, httpserver.StatusServerError, rw.Body.String())
		})
	})

	t.Run("Marshal error", func(t *testing.T) {
		pv := &coremocks.ProtocolVersion{}
		pv.ProtocolReturns(protocol.Protocol{})

		pc := &mocks.ProtocolClient{}
		pc.CurrentReturns(pv, nil)

		h := NewVersionHandler(channel1, handlerCfg, pc)
		require.NotNil(t, h)

		errExpected := fmt.Errorf("injected json marshal error")

		h.jsonMarshal = func(v interface{}) ([]byte, error) { return nil, errExpected }

		rw := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, handlerCfg.BasePath, nil)

		h.Handler()(rw, req)

		require.Equal(t, http.StatusInternalServerError, rw.Result().StatusCode)
		require.Equal(t, httpserver.ContentTypeText, rw.Header().Get(httpserver.ContentTypeHeader))
		require.Equal(t, httpserver.StatusServerError, rw.Body.String())
	})
}

func setTimeParam(time string) func() {
	restore := getParams
	getParams = func(req *http.Request) map[string][]string { return map[string][]string{timeParam: {time}} }

	return func() { getParams = restore }
}
