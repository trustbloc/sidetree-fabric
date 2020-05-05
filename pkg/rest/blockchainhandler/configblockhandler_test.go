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

	"github.com/golang/protobuf/proto"
	cb "github.com/hyperledger/fabric-protos-go/common"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/stretchr/testify/require"
	"github.com/trustbloc/fabric-peer-ext/pkg/mocks"

	"github.com/trustbloc/sidetree-fabric/pkg/httpserver"
	obmocks "github.com/trustbloc/sidetree-fabric/pkg/observer/mocks"
)

const (
	configBlockPath = "/blockchain/config-block"
	blockNum        = uint64(999)
	configBlockNum  = uint64(2)
)

func TestNewConfigBlockHandler(t *testing.T) {
	h := NewConfigBlockHandler(channel1, handlerCfg, &obmocks.BlockchainClientProvider{})
	require.NotNil(t, h)

	require.Equal(t, configBlockPath, h.Path())
	require.Equal(t, http.MethodGet, h.Method())
	require.Empty(t, h.Params())
}

func TestNewConfigBlockHandlerWithEncoding(t *testing.T) {
	h := NewConfigBlockHandlerWithEncoding(channel1, handlerCfg, &obmocks.BlockchainClientProvider{})
	require.NotNil(t, h)

	require.Equal(t, configBlockPath, h.Path())
	require.Equal(t, http.MethodGet, h.Method())
	require.Equal(t, map[string]string{"data-encoding": "{data-encoding}"}, h.Params())
}

func TestConfigBlock_Handler(t *testing.T) {
	bcInfo := &cb.BlockchainInfo{
		Height:           1000,
		CurrentBlockHash: []byte{1, 2, 3, 4},
	}

	bcProvider := &obmocks.BlockchainClientProvider{}

	t.Run("With default (JSON) encoding -> Success", func(t *testing.T) {
		bcClient := &obmocks.BlockchainClient{}
		bcClient.GetBlockchainInfoReturns(bcInfo, nil)
		bcClient.GetBlockByNumberReturnsOnCall(0, generateBlock(t, blockNum, configBlockNum), nil)
		bcClient.GetBlockByNumberReturnsOnCall(1, generateConfigBlock(t, configBlockNum), nil)

		bcProvider.ForChannelReturns(bcClient, nil)

		h := NewConfigBlockHandler(channel1, handlerCfg, bcProvider)
		require.NotNil(t, h)

		rw := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, configBlockPath, nil)

		h.Handler()(rw, req)

		require.Equal(t, http.StatusOK, rw.Result().StatusCode)
		require.Equal(t, httpserver.ContentTypeJSON, rw.Header().Get(httpserver.ContentTypeHeader))

		var blockResp Block
		require.NoError(t, json.Unmarshal(rw.Body.Bytes(), &blockResp))
		require.NotNil(t, blockResp[headerField])
		require.NotNil(t, blockResp[dataField])
	})

	t.Run("With base64 encoding -> Success", func(t *testing.T) {
		restoreParams := setConfigBlockParams("", DataEncodingBase64)
		defer restoreParams()

		bcClient := &obmocks.BlockchainClient{}
		bcClient.GetBlockchainInfoReturns(bcInfo, nil)
		bcClient.GetBlockByNumberReturnsOnCall(0, generateBlock(t, blockNum, configBlockNum), nil)
		bcClient.GetBlockByNumberReturnsOnCall(1, generateConfigBlock(t, configBlockNum), nil)

		bcProvider.ForChannelReturns(bcClient, nil)

		h := NewConfigBlockHandlerWithEncoding(channel1, handlerCfg, bcProvider)
		require.NotNil(t, h)

		rw := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, configBlockPath, nil)

		h.Handler()(rw, req)

		require.Equal(t, http.StatusOK, rw.Result().StatusCode)
		require.Equal(t, httpserver.ContentTypeJSON, rw.Header().Get(httpserver.ContentTypeHeader))

		var block BlockResponse
		require.NoError(t, json.Unmarshal(rw.Body.Bytes(), &block))
		require.NotNil(t, block.Header)
		require.Equal(t, configBlockNum, block.Header.Number)
		require.NotEmpty(t, block.Header.DataHash)

		dataBytes, err := base64.StdEncoding.DecodeString(block.Header.DataHash)
		require.NoError(t, err)
		require.NotEmpty(t, dataBytes)
	})

	t.Run("With base64URL encoding -> Success", func(t *testing.T) {
		restoreParams := setConfigBlockParams("", DataEncodingBase64URL)
		defer restoreParams()

		const blockNum = uint64(999)

		bcClient := &obmocks.BlockchainClient{}
		bcClient.GetBlockchainInfoReturns(bcInfo, nil)
		bcClient.GetBlockByNumberReturnsOnCall(0, generateBlock(t, blockNum, configBlockNum), nil)
		bcClient.GetBlockByNumberReturnsOnCall(1, generateConfigBlock(t, configBlockNum), nil)

		bcProvider.ForChannelReturns(bcClient, nil)

		h := NewConfigBlockHandlerWithEncoding(channel1, handlerCfg, bcProvider)
		require.NotNil(t, h)

		rw := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, configBlockPath, nil)

		h.Handler()(rw, req)

		require.Equal(t, http.StatusOK, rw.Result().StatusCode)
		require.Equal(t, httpserver.ContentTypeJSON, rw.Header().Get(httpserver.ContentTypeHeader))

		var block BlockResponse
		require.NoError(t, json.Unmarshal(rw.Body.Bytes(), &block))
		require.NotNil(t, block.Header)
		require.Equal(t, configBlockNum, block.Header.Number)
		require.NotEmpty(t, block.Header.DataHash)

		dataBytes, err := base64.URLEncoding.DecodeString(block.Header.DataHash)
		require.NoError(t, err)
		require.NotEmpty(t, dataBytes)
	})

	t.Run("Marshal error -> Server Error", func(t *testing.T) {
		bcClient := &obmocks.BlockchainClient{}
		bcClient.GetBlockchainInfoReturns(bcInfo, nil)
		bcClient.GetBlockByNumberReturnsOnCall(0, generateBlock(t, blockNum, configBlockNum), nil)
		bcClient.GetBlockByNumberReturnsOnCall(1, generateConfigBlock(t, configBlockNum), nil)

		bcProvider.ForChannelReturns(bcClient, nil)

		h := NewConfigBlockHandler(channel1, handlerCfg, bcProvider)
		require.NotNil(t, h)
		h.jsonMarshal = func(v interface{}) ([]byte, error) { return nil, errors.New("injected marshal error") }

		rw := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, configBlockPath, nil)

		h.Handler()(rw, req)

		require.Equal(t, httpserver.StatusServerError, rw.Body.String())
		require.Equal(t, http.StatusInternalServerError, rw.Result().StatusCode)
		require.Equal(t, httpserver.ContentTypeText, rw.Header().Get(httpserver.ContentTypeHeader))
	})

	t.Run("Blockchain provider error -> Server Error", func(t *testing.T) {
		bcProvider.ForChannelReturns(nil, errors.New("injected blockchain provider error"))

		h := NewConfigBlockHandler(channel1, handlerCfg, bcProvider)
		require.NotNil(t, h)

		rw := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, configBlockPath, nil)

		h.Handler()(rw, req)

		require.Equal(t, http.StatusInternalServerError, rw.Result().StatusCode)
		require.Equal(t, httpserver.ContentTypeText, rw.Header().Get(httpserver.ContentTypeHeader))
		require.Equal(t, httpserver.StatusServerError, rw.Body.String())
	})

	t.Run("Blockchain info error -> Server Error", func(t *testing.T) {
		bcClient := &obmocks.BlockchainClient{}
		bcClient.GetBlockchainInfoReturns(nil, errors.New("injected blockchain client error"))

		bcProvider.ForChannelReturns(bcClient, nil)

		h := NewConfigBlockHandler(channel1, handlerCfg, bcProvider)
		require.NotNil(t, h)

		rw := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, configBlockPath, nil)

		h.Handler()(rw, req)

		require.Equal(t, http.StatusInternalServerError, rw.Result().StatusCode)
		require.Equal(t, httpserver.ContentTypeText, rw.Header().Get(httpserver.ContentTypeHeader))
		require.Equal(t, httpserver.StatusServerError, rw.Body.String())
	})

	t.Run("BlockByNumber error -> Server Error", func(t *testing.T) {
		bcClient := &obmocks.BlockchainClient{}
		bcClient.GetBlockchainInfoReturns(bcInfo, nil)
		bcClient.GetBlockByNumberReturns(nil, errors.New("injected blockchain client error"))

		bcProvider.ForChannelReturns(bcClient, nil)

		h := NewConfigBlockHandler(channel1, handlerCfg, bcProvider)
		require.NotNil(t, h)

		rw := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, configBlockPath, nil)

		h.Handler()(rw, req)

		require.Equal(t, http.StatusInternalServerError, rw.Result().StatusCode)
		require.Equal(t, httpserver.ContentTypeText, rw.Header().Get(httpserver.ContentTypeHeader))
		require.Equal(t, httpserver.StatusServerError, rw.Body.String())
	})

	t.Run("ConfigBlockByNumber error -> Server Error", func(t *testing.T) {
		bcClient := &obmocks.BlockchainClient{}
		bcClient.GetBlockchainInfoReturns(bcInfo, nil)
		bcClient.GetBlockByNumberReturnsOnCall(0, generateBlock(t, blockNum, configBlockNum), nil)
		bcClient.GetBlockByNumberReturnsOnCall(1, nil, errors.New("injected blockchain client error"))

		bcProvider.ForChannelReturns(bcClient, nil)

		h := NewConfigBlockHandler(channel1, handlerCfg, bcProvider)
		require.NotNil(t, h)

		rw := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, configBlockPath, nil)

		h.Handler()(rw, req)

		require.Equal(t, http.StatusInternalServerError, rw.Result().StatusCode)
		require.Equal(t, httpserver.ContentTypeText, rw.Header().Get(httpserver.ContentTypeHeader))
		require.Equal(t, httpserver.StatusServerError, rw.Body.String())
	})
}

func TestNewConfigBlockByHashHandler(t *testing.T) {
	h := NewConfigBlockByHashHandler(channel1, handlerCfg, &obmocks.BlockchainClientProvider{})
	require.NotNil(t, h)

	require.Equal(t, "/blockchain/config-block/{hash}", h.Path())
	require.Equal(t, http.MethodGet, h.Method())
	require.Empty(t, h.Params())
}

func TestNewConfigBlockByHashHandlerWithEncoding(t *testing.T) {
	h := NewConfigBlockByHashHandlerWithEncoding(channel1, handlerCfg, &obmocks.BlockchainClientProvider{})
	require.NotNil(t, h)

	require.Equal(t, "/blockchain/config-block/{hash}", h.Path())
	require.Equal(t, http.MethodGet, h.Method())
	require.Equal(t, map[string]string{"data-encoding": "{data-encoding}"}, h.Params())
}

func TestConfigBlock_ByHashHandler(t *testing.T) {
	const (
		txn1 = "tx1"
	)

	bcProvider := &obmocks.BlockchainClientProvider{}

	t.Run("With default (JSON) encoding -> Success", func(t *testing.T) {
		restoreParams := setConfigBlockParams(base64.URLEncoding.EncodeToString([]byte("hash1")), "")
		defer restoreParams()

		bcClient := &obmocks.BlockchainClient{}
		bcClient.GetBlockByHashReturns(generateBlock(t, blockNum, configBlockNum), nil)
		bcClient.GetBlockByNumberReturns(generateConfigBlock(t, configBlockNum), nil)

		bcProvider.ForChannelReturns(bcClient, nil)

		h := NewConfigBlockByHashHandler(channel1, handlerCfg, bcProvider)
		require.NotNil(t, h)

		rw := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, configBlockPath, nil)

		h.Handler()(rw, req)

		require.Equal(t, http.StatusOK, rw.Result().StatusCode)
		require.Equal(t, httpserver.ContentTypeJSON, rw.Header().Get(httpserver.ContentTypeHeader))

		var blockResp Block
		require.NoError(t, json.Unmarshal(rw.Body.Bytes(), &blockResp))
		require.NotNil(t, blockResp[headerField])
		require.NotNil(t, blockResp[dataField])
	})

	t.Run("With base64 encoding -> Success", func(t *testing.T) {
		restoreParams := setConfigBlockParams(base64.URLEncoding.EncodeToString([]byte("hash1")), DataEncodingBase64)
		defer restoreParams()

		bcClient := &obmocks.BlockchainClient{}
		bcClient.GetBlockByHashReturns(generateBlock(t, blockNum, configBlockNum), nil)
		bcClient.GetBlockByNumberReturns(generateConfigBlock(t, configBlockNum), nil)

		bcProvider.ForChannelReturns(bcClient, nil)

		h := NewConfigBlockByHashHandlerWithEncoding(channel1, handlerCfg, bcProvider)
		require.NotNil(t, h)

		rw := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, configBlockPath, nil)

		h.Handler()(rw, req)

		require.Equal(t, http.StatusOK, rw.Result().StatusCode)
		require.Equal(t, httpserver.ContentTypeJSON, rw.Header().Get(httpserver.ContentTypeHeader))

		var blockResp BlockResponse
		require.NoError(t, json.Unmarshal(rw.Body.Bytes(), &blockResp))
		require.NotNil(t, blockResp.Header)
		require.Equal(t, configBlockNum, blockResp.Header.Number)
		require.NotEmpty(t, blockResp.Header.DataHash)

		dataBytes, err := base64.StdEncoding.DecodeString(blockResp.Header.DataHash)
		require.NoError(t, err)
		require.NotEmpty(t, dataBytes)
	})

	t.Run("With base64URL encoding -> Success", func(t *testing.T) {
		restoreParams := setConfigBlockParams(base64.URLEncoding.EncodeToString([]byte("hash1")), DataEncodingBase64URL)
		defer restoreParams()

		bcClient := &obmocks.BlockchainClient{}
		bcClient.GetBlockByHashReturns(generateBlock(t, blockNum, configBlockNum), nil)
		bcClient.GetBlockByNumberReturns(generateConfigBlock(t, configBlockNum), nil)

		bcProvider.ForChannelReturns(bcClient, nil)

		h := NewConfigBlockByHashHandlerWithEncoding(channel1, handlerCfg, bcProvider)
		require.NotNil(t, h)

		rw := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, configBlockPath, nil)

		h.Handler()(rw, req)

		require.Equal(t, http.StatusOK, rw.Result().StatusCode)
		require.Equal(t, httpserver.ContentTypeJSON, rw.Header().Get(httpserver.ContentTypeHeader))

		var blockResp BlockResponse
		require.NoError(t, json.Unmarshal(rw.Body.Bytes(), &blockResp))
		require.NotNil(t, blockResp.Header)
		require.Equal(t, configBlockNum, blockResp.Header.Number)
		require.NotEmpty(t, blockResp.Header.DataHash)

		dataBytes, err := base64.URLEncoding.DecodeString(blockResp.Header.DataHash)
		require.NoError(t, err)
		require.NotEmpty(t, dataBytes)
	})

	t.Run("Missing hash param -> BadRequest", func(t *testing.T) {
		testInvalidConfigParams(t, NewConfigBlockByHashHandler, InvalidTimeHash)
	})

	t.Run("Invalid hash param -> BadRequest", func(t *testing.T) {
		restoreParams := setConfigBlockParams("xxx", "")
		defer restoreParams()

		testInvalidConfigParams(t, NewConfigBlockByHashHandler, InvalidTimeHash)
	})

	t.Run("Marshal error -> Server Error", func(t *testing.T) {
		restoreParams := setConfigBlockParams(base64.URLEncoding.EncodeToString([]byte("hash1")), "")
		defer restoreParams()

		bcClient := &obmocks.BlockchainClient{}
		bcClient.GetBlockByHashReturns(generateBlock(t, blockNum, configBlockNum), nil)
		bcClient.GetBlockByNumberReturns(generateConfigBlock(t, configBlockNum), nil)

		bcProvider.ForChannelReturns(bcClient, nil)

		h := NewConfigBlockByHashHandler(channel1, handlerCfg, bcProvider)
		require.NotNil(t, h)
		h.jsonMarshal = func(v interface{}) ([]byte, error) { return nil, errors.New("injected marshal error") }

		rw := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, configBlockPath, nil)

		h.Handler()(rw, req)

		require.Equal(t, httpserver.StatusServerError, rw.Body.String())
		require.Equal(t, http.StatusInternalServerError, rw.Result().StatusCode)
		require.Equal(t, httpserver.ContentTypeText, rw.Header().Get(httpserver.ContentTypeHeader))
	})

	t.Run("Blockchain provider error -> Server Error", func(t *testing.T) {
		restoreParams := setConfigBlockParams(base64.URLEncoding.EncodeToString([]byte("hash1")), "")
		defer restoreParams()

		bcProvider.ForChannelReturns(nil, errors.New("injected blockchain provider error"))

		h := NewConfigBlockByHashHandler(channel1, handlerCfg, bcProvider)
		require.NotNil(t, h)

		rw := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, configBlockPath, nil)

		h.Handler()(rw, req)

		require.Equal(t, http.StatusInternalServerError, rw.Result().StatusCode)
		require.Equal(t, httpserver.ContentTypeText, rw.Header().Get(httpserver.ContentTypeHeader))
		require.Equal(t, httpserver.StatusServerError, rw.Body.String())
	})

	t.Run("Blockchain client error -> Server Error", func(t *testing.T) {
		restoreParams := setConfigBlockParams(base64.URLEncoding.EncodeToString([]byte("hash1")), "")
		defer restoreParams()

		bcClient := &obmocks.BlockchainClient{}
		bcClient.GetBlockByHashReturns(nil, errors.New("injected blockchain client error"))

		bcProvider.ForChannelReturns(bcClient, nil)

		h := NewConfigBlockByHashHandler(channel1, handlerCfg, bcProvider)
		require.NotNil(t, h)

		rw := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, configBlockPath, nil)

		h.Handler()(rw, req)

		require.Equal(t, http.StatusInternalServerError, rw.Result().StatusCode)
		require.Equal(t, httpserver.ContentTypeText, rw.Header().Get(httpserver.ContentTypeHeader))
		require.Equal(t, httpserver.StatusServerError, rw.Body.String())
	})
}

type configHandlerCreator func(channelID string, cfg Config, blockchainProvider blockchainClientProvider) *ConfigBlock

func testInvalidConfigParams(t *testing.T, newHandler configHandlerCreator, expectedCode ResultCode) {
	const (
		txn1 = "tx1"
	)

	bcClient := &obmocks.BlockchainClient{}
	bcClient.GetBlockByHashReturns(generateBlock(t, blockNum, configBlockNum), nil)
	bcClient.GetBlockByNumberReturns(generateConfigBlock(t, configBlockNum), nil)

	bcProvider := &obmocks.BlockchainClientProvider{}
	bcProvider.ForChannelReturns(bcClient, nil)

	rw := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, configBlockPath, nil)

	h := newHandler(channel1, handlerCfg, bcProvider)
	require.NotNil(t, h)

	h.Handler()(rw, req)

	require.Equal(t, http.StatusBadRequest, rw.Result().StatusCode)
	require.Equal(t, httpserver.ContentTypeJSON, rw.Header().Get(httpserver.ContentTypeHeader))

	errResp := &ErrorResponse{}
	require.NoError(t, json.Unmarshal(rw.Body.Bytes(), errResp))
	require.Equal(t, expectedCode, errResp.Code)
}

func setConfigBlockParams(hash string, encoding DataEncoding) func() {
	restoreGetHash := getHash
	restoreGetDataEncoding := getDataEncoding

	getHash = func(_ *http.Request) string {
		return hash
	}

	getDataEncoding = func(_ *http.Request) DataEncoding {
		return encoding
	}

	return func() {
		getHash = restoreGetHash
		getDataEncoding = restoreGetDataEncoding
	}
}

func generateBlock(t *testing.T, blockNum, cfgBlockNum uint64) *cb.Block {
	bb := mocks.NewBlockBuilder(channel1, blockNum)
	bb.Transaction("txn1", pb.TxValidationCode_VALID).ChaincodeAction("sidetree").Write("key1", []byte("value1"))

	block := bb.Build()
	block.Metadata = generateConfigMetaData(t, cfgBlockNum)

	return block
}

func generateConfigBlock(t *testing.T, blockNum uint64) *cb.Block {
	cfgbb := mocks.NewBlockBuilder(channel1, configBlockNum)
	cfgbb.ConfigUpdate()
	return cfgbb.Build()
}

func generateConfigMetaData(t *testing.T, cfgBlockNum uint64) *cb.BlockMetadata {
	obm := &cb.OrdererBlockMetadata{
		LastConfig: &cb.LastConfig{
			Index: cfgBlockNum,
		},
	}

	obmBytes, err := proto.Marshal(obm)
	require.NoError(t, err)

	md := &cb.Metadata{
		Value: obmBytes,
	}

	sigBytes, err := proto.Marshal(md)
	require.NoError(t, err)

	return &cb.BlockMetadata{
		Metadata: [][]byte{sigBytes},
	}
}
