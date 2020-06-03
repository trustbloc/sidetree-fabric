/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package discoveryhandler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/trustbloc/sidetree-fabric/pkg/peer/discovery"

	"github.com/trustbloc/sidetree-fabric/pkg/httpserver"
	"github.com/trustbloc/sidetree-fabric/pkg/rest/discoveryhandler/mocks"
)

//go:generate counterfeiter -o ./mocks/servicesprovider.gen.go --fake-name ServicesProvider . servicesProvider

const (
	channel1 = "channel1"
)

var (
	handlerCfg = Config{
		BasePath: "/discovery",
	}
)

func TestNew(t *testing.T) {
	h := New(channel1, handlerCfg, &mocks.ServicesProvider{})
	require.NotNil(t, h)

	require.Equal(t, "/discovery", h.Path())
	require.Equal(t, http.MethodGet, h.Method())
}

func TestHandler(t *testing.T) {
	allServices := Services{s1_1, s2_2}

	t.Run("All services", func(t *testing.T) {
		provider := &mocks.ServicesProvider{}
		provider.ServicesForChannelReturns(allServices)

		h := New(channel1, handlerCfg, provider)
		require.NotNil(t, h)

		rw := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/discovery", nil)
		h.Handler()(rw, req)

		require.Equal(t, http.StatusOK, rw.Result().StatusCode)
		require.Equal(t, httpserver.ContentTypeJSON, rw.Header().Get(httpserver.ContentTypeHeader))

		var services []discovery.Service
		require.NoError(t, json.Unmarshal(rw.Body.Bytes(), &services))
		require.Len(t, services, 2)
		require.Equal(t, s1_1.Service, services[0].Service)
		require.Equal(t, s2_2.Service, services[1].Service)
	})

	t.Run("Filtered services", func(t *testing.T) {
		provider := &mocks.ServicesProvider{}
		provider.ServicesForChannelReturns(allServices)

		h := New(channel1, handlerCfg, provider)
		require.NotNil(t, h)

		restore := getParams
		getParams = func(req *http.Request) map[string][]string {
			return map[string][]string{serviceFilterParam: {s2_2.Service}}
		}
		defer func() { getParams = restore }()

		rw := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/discovery", nil)
		h.Handler()(rw, req)

		require.Equal(t, http.StatusOK, rw.Result().StatusCode)
		require.Equal(t, httpserver.ContentTypeJSON, rw.Header().Get(httpserver.ContentTypeHeader))

		var services []discovery.Service
		require.NoError(t, json.Unmarshal(rw.Body.Bytes(), &services))
		require.Len(t, services, 1)
		require.Equal(t, s2_2.Service, services[0].Service)
	})

	t.Run("Marshal error", func(t *testing.T) {
		provider := &mocks.ServicesProvider{}
		provider.ServicesForChannelReturns(allServices)

		errExpected := errors.New("json marshal error")
		h := newHandler(channel1, handlerCfg, provider, func(v interface{}) ([]byte, error) {
			return nil, errExpected
		})
		require.NotNil(t, h)

		rw := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/discovery", nil)
		h.Handler()(rw, req)

		require.Equal(t, http.StatusInternalServerError, rw.Result().StatusCode)
		require.Equal(t, httpserver.ContentTypeText, rw.Header().Get(httpserver.ContentTypeHeader))
		require.Equal(t, httpserver.StatusServerError, rw.Body.String())
	})
}
