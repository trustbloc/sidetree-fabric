/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package authhandler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/trustbloc/sidetree-fabric/pkg/rest/authhandler/mocks"
)

//go:generate counterfeiter -o ./mocks/httphandler.gen.go --fake-name HTTPHandler github.com/trustbloc/sidetree-core-go/pkg/restapi/common.HTTPHandler

const (
	channel1 = "channel1"
)

func TestNew(t *testing.T) {
	handler := &mocks.HTTPHandler{}
	handler.HandlerReturns(func(http.ResponseWriter, *http.Request) {})

	t.Run("With no tokens", func(t *testing.T) {
		h := New(channel1, []string{}, handler)
		require.NotNil(t, h)
		require.NotNil(t, h.Handler())
	})

	t.Run("With tokens", func(t *testing.T) {
		h := New(channel1, []string{"t1", "t2"}, handler)
		require.NotNil(t, h)
	})
}

func TestHandler_Handle(t *testing.T) {
	handler := &mocks.HTTPHandler{}
	handler.HandlerReturns(func(http.ResponseWriter, *http.Request) {})

	t.Run("No token in header -> Unauthorized", func(t *testing.T) {
		rw := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/blockchain/info", nil)

		h := New(channel1, []string{"t1", "t2"}, handler)
		require.NotNil(t, h)

		h.Handler()(rw, req)
		require.Equal(t, http.StatusUnauthorized, rw.Code)
	})

	t.Run("Valid token in header -> Authorized", func(t *testing.T) {
		token := "some token"

		rw := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/blockchain/info", nil)
		req.Header.Add(authHeader, tokenPrefix+token)

		h := New(channel1, []string{token}, handler)
		require.NotNil(t, h)

		h.Handler()(rw, req)
		require.Equal(t, http.StatusOK, rw.Code)
	})
}
