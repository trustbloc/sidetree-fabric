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
	"strconv"
	"testing"

	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric/protoutil"
	"github.com/stretchr/testify/require"

	"github.com/trustbloc/sidetree-fabric/pkg/httpserver"
	"github.com/trustbloc/sidetree-fabric/pkg/observer/mocks"
)

const (
	channel1 = "channel1"
	v1       = "1.0.1"
)

var (
	handlerCfg = Config{
		BasePath:                  "/blockchain",
		MaxTransactionsInResponse: 2,
	}
)

func TestNewTimeHandler(t *testing.T) {
	bcProvider := &mocks.BlockchainClientProvider{}

	h := NewTimeHandler(channel1, handlerCfg, bcProvider)
	require.NotNil(t, h)

	require.Equal(t, "/blockchain/time", h.Path())
	require.Equal(t, http.MethodGet, h.Method())
}

func TestTime_Latest(t *testing.T) {
	bcInfo := &common.BlockchainInfo{
		Height:           1000,
		CurrentBlockHash: []byte{1, 2, 3, 4},
	}

	bcProvider := &mocks.BlockchainClientProvider{}

	t.Run("Success", func(t *testing.T) {
		bcClient := &mocks.BlockchainClient{}
		bcClient.GetBlockchainInfoReturns(bcInfo, nil)

		bcProvider.ForChannelReturns(bcClient, nil)

		h := NewTimeHandler(channel1, handlerCfg, bcProvider)
		require.NotNil(t, h)

		rw := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/blockchain/time", nil)

		h.Handler()(rw, req)

		require.Equal(t, http.StatusOK, rw.Result().StatusCode)
		require.Equal(t, httpserver.ContentTypeJSON, rw.Header().Get(httpserver.ContentTypeHeader))

		resp := &TimeResponse{}
		require.NoError(t, json.Unmarshal(rw.Body.Bytes(), resp))
		require.Equal(t, strconv.FormatUint(bcInfo.Height-1, 10), resp.Time)
		require.Equal(t, base64.URLEncoding.EncodeToString(bcInfo.CurrentBlockHash), resp.Hash)
	})

	t.Run("Marshal error -> Server Error", func(t *testing.T) {
		bcClient := &mocks.BlockchainClient{}
		bcClient.GetBlockchainInfoReturns(bcInfo, nil)

		bcProvider.ForChannelReturns(bcClient, nil)

		h := NewTimeHandler(channel1, handlerCfg, bcProvider)
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

		h := NewTimeHandler(channel1, handlerCfg, bcProvider)
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

		h := NewTimeHandler(channel1, handlerCfg, bcProvider)
		require.NotNil(t, h)

		rw := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/blockchain/time", nil)

		h.Handler()(rw, req)

		require.Equal(t, http.StatusInternalServerError, rw.Result().StatusCode)
		require.Equal(t, httpserver.ContentTypeText, rw.Header().Get(httpserver.ContentTypeHeader))
		require.Equal(t, httpserver.StatusServerError, rw.Body.String())
	})
}

func TestTime_ByHash(t *testing.T) {
	block := &common.Block{
		Header: &common.BlockHeader{
			Number:       1000,
			PreviousHash: []byte{1, 2, 3, 4},
			DataHash:     []byte{5, 6, 7, 8},
		},
	}

	bcProvider := &mocks.BlockchainClientProvider{}

	t.Run("Success", func(t *testing.T) {
		restoreHash := setParams("Zi6XcyzoikY-OPOme_l2zQpexGdLov1-23ciPN66QQ8=")
		defer restoreHash()

		bcClient := &mocks.BlockchainClient{}
		bcClient.GetBlockByHashReturns(block, nil)

		bcProvider.ForChannelReturns(bcClient, nil)

		h := NewTimeByHashHandler(channel1, handlerCfg, bcProvider)
		require.NotNil(t, h)

		rw := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/blockchain/time/1234", nil)

		h.Handler()(rw, req)

		require.Equal(t, http.StatusOK, rw.Result().StatusCode)
		require.Equal(t, httpserver.ContentTypeJSON, rw.Header().Get(httpserver.ContentTypeHeader))

		resp := &TimeResponse{}
		require.NoError(t, json.Unmarshal(rw.Body.Bytes(), resp))
		require.Equal(t, strconv.FormatUint(block.Header.Number, 10), resp.Time)
		require.Equal(t, base64.URLEncoding.EncodeToString(protoutil.BlockHeaderHash(block.Header)), resp.Hash)
	})

	t.Run("Not Found", func(t *testing.T) {
		restoreHash := setParams("Zi6XcyzoikY-OPOme_l2zQpexGdLov1-23ciPN66QQ8=")
		defer restoreHash()

		bcClient := &mocks.BlockchainClient{}
		bcClient.GetBlockByHashReturns(nil, errors.New("not found"))

		bcProvider.ForChannelReturns(bcClient, nil)

		h := NewTimeByHashHandler(channel1, handlerCfg, bcProvider)
		require.NotNil(t, h)

		rw := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/blockchain/time/1234", nil)

		h.Handler()(rw, req)

		require.Equal(t, http.StatusNotFound, rw.Result().StatusCode)
		require.Equal(t, httpserver.ContentTypeText, rw.Header().Get(httpserver.ContentTypeHeader))
		require.Equal(t, httpserver.StatusNotFound, rw.Body.String())
	})

	t.Run("Bad block -> Server Error", func(t *testing.T) {
		restoreHash := setParams("Zi6XcyzoikY-OPOme_l2zQpexGdLov1-23ciPN66QQ8=")
		defer restoreHash()

		bcClient := &mocks.BlockchainClient{}
		bcClient.GetBlockByHashReturns(&common.Block{}, nil)

		bcProvider.ForChannelReturns(bcClient, nil)

		h := NewTimeByHashHandler(channel1, handlerCfg, bcProvider)
		require.NotNil(t, h)

		rw := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/blockchain/time/1234", nil)

		h.Handler()(rw, req)

		require.Equal(t, http.StatusInternalServerError, rw.Result().StatusCode)
		require.Equal(t, httpserver.ContentTypeText, rw.Header().Get(httpserver.ContentTypeHeader))
		require.Equal(t, httpserver.StatusServerError, rw.Body.String())
	})

	t.Run("Missing hash -> Bad Request", func(t *testing.T) {
		h := NewTimeByHashHandler(channel1, handlerCfg, bcProvider)
		require.NotNil(t, h)

		rw := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/blockchain/time/1234", nil)

		h.Handler()(rw, req)

		require.Equal(t, http.StatusBadRequest, rw.Result().StatusCode)
		require.Equal(t, httpserver.ContentTypeText, rw.Header().Get(httpserver.ContentTypeHeader))
		require.Equal(t, httpserver.StatusBadRequest, rw.Body.String())
	})

	t.Run("Invalid hash -> Bad Request", func(t *testing.T) {
		restoreHash := setParams("xxx_xxx")
		defer restoreHash()

		h := NewTimeByHashHandler(channel1, handlerCfg, bcProvider)
		require.NotNil(t, h)

		rw := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/blockchain/time/1234", nil)

		h.Handler()(rw, req)

		require.Equal(t, http.StatusBadRequest, rw.Result().StatusCode)
		require.Equal(t, httpserver.ContentTypeText, rw.Header().Get(httpserver.ContentTypeHeader))
		require.Equal(t, httpserver.StatusBadRequest, rw.Body.String())
	})

	t.Run("Marshal error -> Server Error", func(t *testing.T) {
		restoreHash := setParams("Zi6XcyzoikY-OPOme_l2zQpexGdLov1-23ciPN66QQ8=")
		defer restoreHash()

		bcClient := &mocks.BlockchainClient{}
		bcClient.GetBlockByHashReturns(block, nil)

		bcProvider.ForChannelReturns(bcClient, nil)

		h := NewTimeByHashHandler(channel1, handlerCfg, bcProvider)
		require.NotNil(t, h)
		h.jsonMarshal = func(v interface{}) ([]byte, error) { return nil, errors.New("injected marshal error") }

		rw := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/blockchain/time/1234", nil)

		h.Handler()(rw, req)

		require.Equal(t, http.StatusInternalServerError, rw.Result().StatusCode)
		require.Equal(t, httpserver.ContentTypeText, rw.Header().Get(httpserver.ContentTypeHeader))
		require.Equal(t, httpserver.StatusServerError, rw.Body.String())
	})

	t.Run("Blockchain provider error -> Server Error", func(t *testing.T) {
		restoreHash := setParams("Zi6XcyzoikY-OPOme_l2zQpexGdLov1-23ciPN66QQ8=")
		defer restoreHash()

		bcProvider.ForChannelReturns(nil, errors.New("injected blockchain provider error"))

		h := NewTimeByHashHandler(channel1, handlerCfg, bcProvider)
		require.NotNil(t, h)

		rw := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/blockchain/time/1234", nil)

		h.Handler()(rw, req)

		require.Equal(t, http.StatusInternalServerError, rw.Result().StatusCode)
		require.Equal(t, httpserver.ContentTypeText, rw.Header().Get(httpserver.ContentTypeHeader))
		require.Equal(t, httpserver.StatusServerError, rw.Body.String())
	})

	t.Run("Blockchain client error -> Server Error", func(t *testing.T) {
		restoreHash := setParams("Zi6XcyzoikY-OPOme_l2zQpexGdLov1-23ciPN66QQ8=")
		defer restoreHash()

		bcClient := &mocks.BlockchainClient{}
		bcClient.GetBlockByHashReturns(nil, errors.New("injected blockchain client error"))

		bcProvider.ForChannelReturns(bcClient, nil)

		h := NewTimeByHashHandler(channel1, handlerCfg, bcProvider)
		require.NotNil(t, h)

		rw := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/blockchain/time/1234", nil)

		h.Handler()(rw, req)

		require.Equal(t, http.StatusInternalServerError, rw.Result().StatusCode)
		require.Equal(t, httpserver.ContentTypeText, rw.Header().Get(httpserver.ContentTypeHeader))
		require.Equal(t, httpserver.StatusServerError, rw.Body.String())
	})
}

func setParams(hash string) func() {
	restoreHash := getHash

	getHash = func(req *http.Request) string { return hash }

	return func() {
		getHash = restoreHash
	}
}
