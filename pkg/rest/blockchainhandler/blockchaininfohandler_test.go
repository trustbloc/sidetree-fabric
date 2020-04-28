/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package blockchainhandler

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/stretchr/testify/require"

	"github.com/trustbloc/sidetree-fabric/pkg/httpserver"
	"github.com/trustbloc/sidetree-fabric/pkg/observer/mocks"
)

func TestNewInfoHandler(t *testing.T) {
	bcProvider := &mocks.BlockchainClientProvider{}

	h := NewInfoHandler(channel1, handlerCfg, bcProvider)
	require.NotNil(t, h)

	require.Equal(t, "/blockchain/info", h.Path())
	require.Equal(t, http.MethodGet, h.Method())
}

func TestInfo_Latest(t *testing.T) {
	bcInfo := &common.BlockchainInfo{
		Height:           1000,
		CurrentBlockHash: []byte{1, 2, 3, 4},
	}

	bcProvider := &mocks.BlockchainClientProvider{}

	t.Run("Success", func(t *testing.T) {
		bcClient := &mocks.BlockchainClient{}
		bcClient.GetBlockchainInfoReturns(bcInfo, nil)

		bcProvider.ForChannelReturns(bcClient, nil)

		h := NewInfoHandler(channel1, handlerCfg, bcProvider)
		require.NotNil(t, h)

		rw := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/blockchain/info", nil)

		h.Handler()(rw, req)

		require.Equal(t, http.StatusOK, rw.Result().StatusCode)
		require.Equal(t, httpserver.ContentTypeJSON, rw.Header().Get(httpserver.ContentTypeHeader))

		resp := &InfoResponse{}
		require.NoError(t, json.Unmarshal(rw.Body.Bytes(), resp))
		require.Equal(t, bcInfo.Height-1, resp.CurrentTime)
		require.Equal(t, base64.URLEncoding.EncodeToString(bcInfo.CurrentBlockHash), resp.CurrentTimeHash)
		require.Equal(t, base64.URLEncoding.EncodeToString(bcInfo.PreviousBlockHash), resp.PreviousTimeHash)
	})

	t.Run("Marshal error -> Server Error", func(t *testing.T) {
		bcClient := &mocks.BlockchainClient{}
		bcClient.GetBlockchainInfoReturns(bcInfo, nil)

		bcProvider.ForChannelReturns(bcClient, nil)

		h := NewInfoHandler(channel1, handlerCfg, bcProvider)
		require.NotNil(t, h)
		h.jsonMarshal = func(v interface{}) ([]byte, error) { return nil, errors.New("injected marshal error") }

		rw := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/blockchain/info", nil)

		h.Handler()(rw, req)

		require.Equal(t, http.StatusInternalServerError, rw.Result().StatusCode)
		require.Equal(t, httpserver.ContentTypeText, rw.Header().Get(httpserver.ContentTypeHeader))
		require.Equal(t, httpserver.StatusServerError, rw.Body.String())
	})

	t.Run("Blockchain provider error -> Server Error", func(t *testing.T) {
		bcProvider.ForChannelReturns(nil, errors.New("injected blockchain provider error"))

		h := NewInfoHandler(channel1, handlerCfg, bcProvider)
		require.NotNil(t, h)

		rw := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/blockchain/info", nil)

		h.Handler()(rw, req)

		require.Equal(t, http.StatusInternalServerError, rw.Result().StatusCode)
		require.Equal(t, httpserver.ContentTypeText, rw.Header().Get(httpserver.ContentTypeHeader))
		require.Equal(t, httpserver.StatusServerError, rw.Body.String())
	})

	t.Run("Blockchain client error -> Server Error", func(t *testing.T) {
		bcClient := &mocks.BlockchainClient{}
		bcClient.GetBlockchainInfoReturns(nil, errors.New("injected blockchain client error"))

		bcProvider.ForChannelReturns(bcClient, nil)

		h := NewInfoHandler(channel1, handlerCfg, bcProvider)
		require.NotNil(t, h)

		rw := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/blockchain/info", nil)

		h.Handler()(rw, req)

		require.Equal(t, http.StatusInternalServerError, rw.Result().StatusCode)
		require.Equal(t, httpserver.ContentTypeText, rw.Header().Get(httpserver.ContentTypeHeader))
		require.Equal(t, httpserver.StatusServerError, rw.Body.String())
	})
}
