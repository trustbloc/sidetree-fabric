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
	"strconv"
	"testing"

	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/stretchr/testify/require"

	"github.com/trustbloc/sidetree-fabric/pkg/httpserver"
	"github.com/trustbloc/sidetree-fabric/pkg/observer/mocks"
)

const (
	channel1 = "channel1"
)

var (
	handlerCfg = Config{
		BasePath: "/blockchain",
	}
)

func TestNewTimeHandler(t *testing.T) {
	bcProvider := &mocks.BlockchainClientProvider{}

	h := NewTimeHandler(channel1, handlerCfg, bcProvider)
	require.NotNil(t, h)

	require.Equal(t, "/blockchain/time", h.Path())
	require.Equal(t, http.MethodGet, h.Method())
}

func TestTime_Handler(t *testing.T) {
	const v1 = "1.0.1"

	cfg := Config{
		BasePath: "/cas",
		Version:  v1,
	}

	bcInfo := &common.BlockchainInfo{
		Height:           1000,
		CurrentBlockHash: []byte{1, 2, 3, 4},
	}

	bcProvider := &mocks.BlockchainClientProvider{}

	t.Run("Success", func(t *testing.T) {
		bcClient := &mocks.BlockchainClient{}
		bcClient.GetBlockchainInfoReturns(bcInfo, nil)

		bcProvider.ForChannelReturns(bcClient, nil)

		h := NewTimeHandler(channel1, cfg, bcProvider)
		require.NotNil(t, h)

		rw := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/blockchain/time", nil)

		h.Handler()(rw, req)

		require.Equal(t, http.StatusOK, rw.Result().StatusCode)
		require.Equal(t, httpserver.ContentTypeJSON, rw.Header().Get(httpserver.ContentTypeHeader))

		resp := &TimeResponse{}
		require.NoError(t, json.Unmarshal(rw.Body.Bytes(), resp))
		require.Equal(t, strconv.FormatUint(bcInfo.Height, 10), resp.Time)
		require.Equal(t, bcInfo.CurrentBlockHash, resp.Hash)
	})

	t.Run("Marshal error -> Server Error", func(t *testing.T) {
		bcClient := &mocks.BlockchainClient{}
		bcClient.GetBlockchainInfoReturns(bcInfo, nil)

		bcProvider.ForChannelReturns(bcClient, nil)

		h := NewTimeHandler(channel1, cfg, bcProvider)
		require.NotNil(t, h)
		h.jsonMarshal = func(v interface{}) ([]byte, error) { return nil, errors.New("injected marshal error") }

		rw := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/blockchain/time", nil)

		h.Handler()(rw, req)

		require.Equal(t, http.StatusInternalServerError, rw.Result().StatusCode)
		require.Equal(t, httpserver.ContentTypeText, rw.Header().Get(httpserver.ContentTypeHeader))
		require.Equal(t, httpserver.StatusServerError, rw.Body.String())
	})

	t.Run("Blockchain provider error -> Server Error", func(t *testing.T) {
		bcProvider.ForChannelReturns(nil, errors.New("injected blockchain provider error"))

		h := NewTimeHandler(channel1, cfg, bcProvider)
		require.NotNil(t, h)

		rw := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/blockchain/time", nil)

		h.Handler()(rw, req)

		require.Equal(t, http.StatusInternalServerError, rw.Result().StatusCode)
		require.Equal(t, httpserver.ContentTypeText, rw.Header().Get(httpserver.ContentTypeHeader))
		require.Equal(t, httpserver.StatusServerError, rw.Body.String())
	})

	t.Run("Blockchain client error -> Server Error", func(t *testing.T) {
		bcClient := &mocks.BlockchainClient{}
		bcClient.GetBlockchainInfoReturns(nil, errors.New("injected blockchain client error"))

		bcProvider.ForChannelReturns(bcClient, nil)
		bcProvider.ForChannelReturns(bcClient, nil)

		h := NewTimeHandler(channel1, cfg, bcProvider)
		require.NotNil(t, h)

		rw := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/blockchain/time", nil)

		h.Handler()(rw, req)

		require.Equal(t, http.StatusInternalServerError, rw.Result().StatusCode)
		require.Equal(t, httpserver.ContentTypeText, rw.Header().Get(httpserver.ContentTypeHeader))
		require.Equal(t, httpserver.StatusServerError, rw.Body.String())
	})
}
