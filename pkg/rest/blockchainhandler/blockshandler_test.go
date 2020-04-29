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
	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/stretchr/testify/require"
	"github.com/trustbloc/fabric-peer-ext/pkg/mocks"

	"github.com/trustbloc/sidetree-fabric/pkg/httpserver"
	obmocks "github.com/trustbloc/sidetree-fabric/pkg/observer/mocks"
)

func TestNewBlocksFromNumHandler(t *testing.T) {
	bcProvider := &obmocks.BlockchainClientProvider{}

	h := NewBlocksFromNumHandler(channel1, handlerCfg, bcProvider)
	require.NotNil(t, h)

	require.Equal(t, "/blockchain/blocks", h.Path())
	require.Equal(t, http.MethodGet, h.Method())
}

func TestBlocks_FromNum(t *testing.T) {
	const (
		txn1 = "tx1"
	)

	bcInfo := &common.BlockchainInfo{
		Height:           1000,
		CurrentBlockHash: []byte{1, 2, 3, 4},
	}

	bcProvider := &obmocks.BlockchainClientProvider{}

	t.Run("With default (JSON) encoding -> Success", func(t *testing.T) {
		restoreParams := setBlocksFromParams("999", "10000")
		defer restoreParams()

		bb := mocks.NewBlockBuilder(channel1, 999)
		bb.Transaction(txn1, pb.TxValidationCode_VALID).ChaincodeAction("sidetree").Write("key1", []byte("value1"))

		bcClient := &obmocks.BlockchainClient{}
		bcClient.GetBlockchainInfoReturns(bcInfo, nil)
		bcClient.GetBlockByNumberReturns(bb.Build(), nil)

		bcProvider.ForChannelReturns(bcClient, nil)

		h := NewBlocksFromNumHandler(channel1, handlerCfg, bcProvider)
		require.NotNil(t, h)

		rw := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/blockchain/blocks", nil)

		h.Handler()(rw, req)

		require.Equal(t, http.StatusOK, rw.Result().StatusCode)
		require.Equal(t, httpserver.ContentTypeJSON, rw.Header().Get(httpserver.ContentTypeHeader))

		var blocks []Block
		require.NoError(t, json.Unmarshal(rw.Body.Bytes(), &blocks))
		require.Len(t, blocks, 1)
		require.NotNil(t, blocks[0][headerField])
		require.NotNil(t, blocks[0][dataField])
	})

	t.Run("With base64 encoding -> Success", func(t *testing.T) {
		const blockNum = uint64(999)

		restoreParams := setBlocksFromWithEncodingParams("999", "1", DataEncodingBase64)
		defer restoreParams()

		bb := mocks.NewBlockBuilder(channel1, blockNum)
		bb.Transaction(txn1, pb.TxValidationCode_VALID).ChaincodeAction("sidetree").Write("key1", []byte("value1"))

		bcClient := &obmocks.BlockchainClient{}
		bcClient.GetBlockchainInfoReturns(bcInfo, nil)
		bcClient.GetBlockByNumberReturns(bb.Build(), nil)

		bcProvider.ForChannelReturns(bcClient, nil)

		h := NewBlocksFromNumHandlerWithEncoding(channel1, handlerCfg, bcProvider)
		require.NotNil(t, h)

		rw := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/blockchain/blocks", nil)

		h.Handler()(rw, req)

		require.Equal(t, http.StatusOK, rw.Result().StatusCode)
		require.Equal(t, httpserver.ContentTypeJSON, rw.Header().Get(httpserver.ContentTypeHeader))

		var blocks []BlockResponse
		require.NoError(t, json.Unmarshal(rw.Body.Bytes(), &blocks))
		require.Len(t, blocks, 1)
		require.NotNil(t, blocks[0].Header)
		require.Equal(t, blockNum, blocks[0].Header.Number)
		require.NotEmpty(t, blocks[0].Header.DataHash)

		dataBytes, err := base64.StdEncoding.DecodeString(blocks[0].Header.DataHash)
		require.NoError(t, err)
		require.NotEmpty(t, dataBytes)
	})

	t.Run("With base64URL encoding -> Success", func(t *testing.T) {
		const blockNum = uint64(999)

		restoreParams := setBlocksFromWithEncodingParams("999", "1", DataEncodingBase64URL)
		defer restoreParams()

		bb := mocks.NewBlockBuilder(channel1, blockNum)
		bb.Transaction(txn1, pb.TxValidationCode_VALID).ChaincodeAction("sidetree").Write("key1", []byte("value1"))

		bcClient := &obmocks.BlockchainClient{}
		bcClient.GetBlockchainInfoReturns(bcInfo, nil)
		bcClient.GetBlockByNumberReturns(bb.Build(), nil)

		bcProvider.ForChannelReturns(bcClient, nil)

		h := NewBlocksFromNumHandlerWithEncoding(channel1, handlerCfg, bcProvider)
		require.NotNil(t, h)

		rw := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/blockchain/blocks", nil)

		h.Handler()(rw, req)

		require.Equal(t, http.StatusOK, rw.Result().StatusCode)
		require.Equal(t, httpserver.ContentTypeJSON, rw.Header().Get(httpserver.ContentTypeHeader))

		var blocks []BlockResponse
		require.NoError(t, json.Unmarshal(rw.Body.Bytes(), &blocks))
		require.Len(t, blocks, 1)
		require.NotNil(t, blocks[0].Header)
		require.Equal(t, blockNum, blocks[0].Header.Number)
		require.NotEmpty(t, blocks[0].Header.DataHash)

		dataBytes, err := base64.URLEncoding.DecodeString(blocks[0].Header.DataHash)
		require.NoError(t, err)
		require.NotEmpty(t, dataBytes)
	})

	t.Run("Missing from-time param -> BadRequest", func(t *testing.T) {
		restoreParams := setBlocksFromParams("", "1")
		defer restoreParams()

		testInvalidParams(t, NewBlocksFromNumHandler, InvalidFromTime)
	})

	t.Run("Invalid from-time param -> BadRequest", func(t *testing.T) {
		restoreParams := setBlocksFromParams("xxx", "1")
		defer restoreParams()

		testInvalidParams(t, NewBlocksFromNumHandler, InvalidFromTime)
	})

	t.Run("Missing max-blocks param -> BadRequest", func(t *testing.T) {
		restoreParams := setBlocksFromParams("999", "")
		defer restoreParams()

		testInvalidParams(t, NewBlocksFromNumHandler, InvalidMaxBlocks)
	})

	t.Run("Invalid max-blocks param -> BadRequest", func(t *testing.T) {
		restoreParams := setBlocksFromParams("999", "xxx")
		defer restoreParams()

		testInvalidParams(t, NewBlocksFromNumHandler, InvalidMaxBlocks)
	})

	t.Run("From-time too big -> NotFound", func(t *testing.T) {
		restoreParams := setBlocksFromParams("10000", "1")
		defer restoreParams()

		bb := mocks.NewBlockBuilder(channel1, 999)
		bb.Transaction(txn1, pb.TxValidationCode_VALID).ChaincodeAction("sidetree").Write("key1", []byte("value1"))

		bcClient := &obmocks.BlockchainClient{}
		bcClient.GetBlockchainInfoReturns(bcInfo, nil)
		bcClient.GetBlockByNumberReturns(bb.Build(), nil)

		bcProvider.ForChannelReturns(bcClient, nil)

		h := NewBlocksFromNumHandler(channel1, handlerCfg, bcProvider)
		require.NotNil(t, h)

		rw := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/blockchain/blocks", nil)

		h.Handler()(rw, req)

		require.Equal(t, http.StatusNotFound, rw.Result().StatusCode)
		require.Equal(t, httpserver.ContentTypeText, rw.Header().Get(httpserver.ContentTypeHeader))
	})

	t.Run("Marshal error -> Server Error", func(t *testing.T) {
		restoreParams := setBlocksFromParams("999", "1")
		defer restoreParams()

		bb := mocks.NewBlockBuilder(channel1, 999)
		bb.Transaction(txn1, pb.TxValidationCode_VALID).ChaincodeAction("sidetree").Write("key1", []byte("value1"))

		bcClient := &obmocks.BlockchainClient{}
		bcClient.GetBlockchainInfoReturns(bcInfo, nil)
		bcClient.GetBlockByNumberReturns(bb.Build(), nil)

		bcProvider.ForChannelReturns(bcClient, nil)

		h := NewBlocksFromNumHandler(channel1, handlerCfg, bcProvider)
		require.NotNil(t, h)
		h.jsonMarshal = func(v interface{}) ([]byte, error) { return nil, errors.New("injected marshal error") }

		rw := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/blockchain/blocks", nil)

		h.Handler()(rw, req)

		require.Equal(t, httpserver.StatusServerError, rw.Body.String())
		require.Equal(t, http.StatusInternalServerError, rw.Result().StatusCode)
		require.Equal(t, httpserver.ContentTypeText, rw.Header().Get(httpserver.ContentTypeHeader))
	})

	t.Run("Blockchain provider error -> Server Error", func(t *testing.T) {
		restoreParams := setBlocksFromParams("999", "1")
		defer restoreParams()

		bcProvider.ForChannelReturns(nil, errors.New("injected blockchain provider error"))

		h := NewBlocksFromNumHandler(channel1, handlerCfg, bcProvider)
		require.NotNil(t, h)

		rw := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/blockchain/blocks", nil)

		h.Handler()(rw, req)

		require.Equal(t, http.StatusInternalServerError, rw.Result().StatusCode)
		require.Equal(t, httpserver.ContentTypeText, rw.Header().Get(httpserver.ContentTypeHeader))
		require.Equal(t, httpserver.StatusServerError, rw.Body.String())
	})

	t.Run("Blockchain client error -> Server Error", func(t *testing.T) {
		restoreParams := setBlocksFromParams("999", "1")
		defer restoreParams()

		bcClient := &obmocks.BlockchainClient{}
		bcClient.GetBlockchainInfoReturns(nil, errors.New("injected blockchain client error"))

		bcProvider.ForChannelReturns(bcClient, nil)

		h := NewBlocksFromNumHandler(channel1, handlerCfg, bcProvider)
		require.NotNil(t, h)

		rw := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/blockchain/blocks", nil)

		h.Handler()(rw, req)

		require.Equal(t, http.StatusInternalServerError, rw.Result().StatusCode)
		require.Equal(t, httpserver.ContentTypeText, rw.Header().Get(httpserver.ContentTypeHeader))
		require.Equal(t, httpserver.StatusServerError, rw.Body.String())
	})
}

func TestNewBlockByHashHandler(t *testing.T) {
	bcProvider := &obmocks.BlockchainClientProvider{}

	h := NewBlockByHashHandler(channel1, handlerCfg, bcProvider)
	require.NotNil(t, h)

	require.Equal(t, "/blockchain/blocks/{hash}", h.Path())
	require.Equal(t, http.MethodGet, h.Method())
}

func TestBlocks_ByHash(t *testing.T) {
	const (
		txn1 = "tx1"
	)

	bcInfo := &common.BlockchainInfo{
		Height:           1000,
		CurrentBlockHash: []byte{1, 2, 3, 4},
	}

	bcProvider := &obmocks.BlockchainClientProvider{}

	t.Run("With default (JSON) encoding -> Success", func(t *testing.T) {
		hash := base64.URLEncoding.EncodeToString([]byte("hash1"))
		restoreParams := setBlocksByHashParams(hash)
		defer restoreParams()

		bb := mocks.NewBlockBuilder(channel1, 999)
		bb.Transaction(txn1, pb.TxValidationCode_VALID).ChaincodeAction("sidetree").Write("key1", []byte("value1"))

		bcClient := &obmocks.BlockchainClient{}
		bcClient.GetBlockchainInfoReturns(bcInfo, nil)
		bcClient.GetBlockByHashReturns(bb.Build(), nil)

		bcProvider.ForChannelReturns(bcClient, nil)

		h := NewBlockByHashHandler(channel1, handlerCfg, bcProvider)
		require.NotNil(t, h)

		rw := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/blockchain/blocks", nil)

		h.Handler()(rw, req)

		require.Equal(t, http.StatusOK, rw.Result().StatusCode)
		require.Equal(t, httpserver.ContentTypeJSON, rw.Header().Get(httpserver.ContentTypeHeader))

		var blocks []Block
		require.NoError(t, json.Unmarshal(rw.Body.Bytes(), &blocks))
		require.Len(t, blocks, 1)
		require.NotNil(t, blocks[0][headerField])
		require.NotNil(t, blocks[0][dataField])
	})

	t.Run("With base64 encoding -> Success", func(t *testing.T) {
		const blockNum = uint64(999)

		hash := base64.URLEncoding.EncodeToString([]byte("hash1"))
		restoreParams := setBlocksByHashWithEncodingParams(hash, DataEncodingBase64)
		defer restoreParams()

		bb := mocks.NewBlockBuilder(channel1, blockNum)
		bb.Transaction(txn1, pb.TxValidationCode_VALID).ChaincodeAction("sidetree").Write("key1", []byte("value1"))

		bcClient := &obmocks.BlockchainClient{}
		bcClient.GetBlockchainInfoReturns(bcInfo, nil)
		bcClient.GetBlockByHashReturns(bb.Build(), nil)

		bcProvider.ForChannelReturns(bcClient, nil)

		h := NewBlockByHashHandlerWithEncoding(channel1, handlerCfg, bcProvider)
		require.NotNil(t, h)

		rw := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/blockchain/blocks", nil)

		h.Handler()(rw, req)

		require.Equal(t, http.StatusOK, rw.Result().StatusCode)
		require.Equal(t, httpserver.ContentTypeJSON, rw.Header().Get(httpserver.ContentTypeHeader))

		var blocks []BlockResponse
		require.NoError(t, json.Unmarshal(rw.Body.Bytes(), &blocks))
		require.Len(t, blocks, 1)
		require.NotNil(t, blocks[0].Header)
		require.Equal(t, blockNum, blocks[0].Header.Number)
		require.NotEmpty(t, blocks[0].Header.DataHash)

		dataBytes, err := base64.StdEncoding.DecodeString(blocks[0].Header.DataHash)
		require.NoError(t, err)
		require.NotEmpty(t, dataBytes)
	})

	t.Run("With base64URL encoding -> Success", func(t *testing.T) {
		const blockNum = uint64(999)

		hash := base64.URLEncoding.EncodeToString([]byte("hash1"))
		restoreParams := setBlocksByHashWithEncodingParams(hash, DataEncodingBase64URL)
		defer restoreParams()

		bb := mocks.NewBlockBuilder(channel1, blockNum)
		bb.Transaction(txn1, pb.TxValidationCode_VALID).ChaincodeAction("sidetree").Write("key1", []byte("value1"))

		bcClient := &obmocks.BlockchainClient{}
		bcClient.GetBlockchainInfoReturns(bcInfo, nil)
		bcClient.GetBlockByHashReturns(bb.Build(), nil)

		bcProvider.ForChannelReturns(bcClient, nil)

		h := NewBlockByHashHandlerWithEncoding(channel1, handlerCfg, bcProvider)
		require.NotNil(t, h)

		rw := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/blockchain/blocks", nil)

		h.Handler()(rw, req)

		require.Equal(t, http.StatusOK, rw.Result().StatusCode)
		require.Equal(t, httpserver.ContentTypeJSON, rw.Header().Get(httpserver.ContentTypeHeader))

		var blocks []BlockResponse
		require.NoError(t, json.Unmarshal(rw.Body.Bytes(), &blocks))
		require.Len(t, blocks, 1)
		require.NotNil(t, blocks[0].Header)
		require.Equal(t, blockNum, blocks[0].Header.Number)
		require.NotEmpty(t, blocks[0].Header.DataHash)

		dataBytes, err := base64.URLEncoding.DecodeString(blocks[0].Header.DataHash)
		require.NoError(t, err)
		require.NotEmpty(t, dataBytes)
	})

	t.Run("Missing hash param -> BadRequest", func(t *testing.T) {
		restoreParams := setBlocksByHashParams("")
		defer restoreParams()

		testInvalidParams(t, NewBlockByHashHandler, InvalidTimeHash)
	})

	t.Run("Invalid hash param -> BadRequest", func(t *testing.T) {
		restoreParams := setBlocksByHashParams("xxx")
		defer restoreParams()

		testInvalidParams(t, NewBlockByHashHandler, InvalidTimeHash)
	})

	t.Run("Marshal error -> Server Error", func(t *testing.T) {
		hash := base64.URLEncoding.EncodeToString([]byte("hash1"))
		restoreParams := setBlocksByHashParams(hash)
		defer restoreParams()

		bb := mocks.NewBlockBuilder(channel1, 999)
		bb.Transaction(txn1, pb.TxValidationCode_VALID).ChaincodeAction("sidetree").Write("key1", []byte("value1"))

		bcClient := &obmocks.BlockchainClient{}
		bcClient.GetBlockchainInfoReturns(bcInfo, nil)
		bcClient.GetBlockByHashReturns(bb.Build(), nil)

		bcProvider.ForChannelReturns(bcClient, nil)

		h := NewBlockByHashHandler(channel1, handlerCfg, bcProvider)
		require.NotNil(t, h)
		h.jsonMarshal = func(v interface{}) ([]byte, error) { return nil, errors.New("injected marshal error") }

		rw := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/blockchain/blocks", nil)

		h.Handler()(rw, req)

		require.Equal(t, httpserver.StatusServerError, rw.Body.String())
		require.Equal(t, http.StatusInternalServerError, rw.Result().StatusCode)
		require.Equal(t, httpserver.ContentTypeText, rw.Header().Get(httpserver.ContentTypeHeader))
	})

	t.Run("Blockchain provider error -> Server Error", func(t *testing.T) {
		hash := base64.URLEncoding.EncodeToString([]byte("hash1"))
		restoreParams := setBlocksByHashParams(hash)
		defer restoreParams()

		bcProvider.ForChannelReturns(nil, errors.New("injected blockchain provider error"))

		h := NewBlockByHashHandler(channel1, handlerCfg, bcProvider)
		require.NotNil(t, h)

		rw := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/blockchain/blocks", nil)

		h.Handler()(rw, req)

		require.Equal(t, http.StatusInternalServerError, rw.Result().StatusCode)
		require.Equal(t, httpserver.ContentTypeText, rw.Header().Get(httpserver.ContentTypeHeader))
		require.Equal(t, httpserver.StatusServerError, rw.Body.String())
	})

	t.Run("Blockchain client error -> Server Error", func(t *testing.T) {
		hash := base64.URLEncoding.EncodeToString([]byte("hash1"))
		restoreParams := setBlocksByHashParams(hash)
		defer restoreParams()

		bcClient := &obmocks.BlockchainClient{}
		bcClient.GetBlockByHashReturns(nil, errors.New("injected blockchain client error"))

		bcProvider.ForChannelReturns(bcClient, nil)

		h := NewBlockByHashHandler(channel1, handlerCfg, bcProvider)
		require.NotNil(t, h)

		rw := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/blockchain/blocks", nil)

		h.Handler()(rw, req)

		require.Equal(t, http.StatusInternalServerError, rw.Result().StatusCode)
		require.Equal(t, httpserver.ContentTypeText, rw.Header().Get(httpserver.ContentTypeHeader))
		require.Equal(t, httpserver.StatusServerError, rw.Body.String())
	})
}

type handlerCreator func(channelID string, cfg Config, blockchainProvider blockchainClientProvider) *Blocks

func testInvalidParams(t *testing.T, newHandler handlerCreator, expectedCode ResultCode) {
	const (
		txn1 = "tx1"
	)

	bcInfo := &common.BlockchainInfo{
		Height:           1000,
		CurrentBlockHash: []byte{1, 2, 3, 4},
	}

	bb := mocks.NewBlockBuilder(channel1, 999)
	bb.Transaction(txn1, pb.TxValidationCode_VALID).ChaincodeAction("sidetree").Write("key1", []byte("value1"))

	bcClient := &obmocks.BlockchainClient{}
	bcClient.GetBlockchainInfoReturns(bcInfo, nil)
	bcClient.GetBlockByNumberReturns(bb.Build(), nil)

	bcProvider := &obmocks.BlockchainClientProvider{}
	bcProvider.ForChannelReturns(bcClient, nil)

	rw := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/blockchain/blocks", nil)

	h := newHandler(channel1, handlerCfg, bcProvider)
	require.NotNil(t, h)

	h.Handler()(rw, req)

	require.Equal(t, http.StatusBadRequest, rw.Result().StatusCode)
	require.Equal(t, httpserver.ContentTypeJSON, rw.Header().Get(httpserver.ContentTypeHeader))

	errResp := &ErrorResponse{}
	require.NoError(t, json.Unmarshal(rw.Body.Bytes(), errResp))
	require.Equal(t, expectedCode, errResp.Code)
}

func setBlocksFromWithEncodingParams(from, maxBlocks string, encoding DataEncoding) func() {
	restoreParams := setBlocksFromParams(from, maxBlocks)
	restoreGetDataEncoding := getDataEncoding

	getDataEncoding = func(_ *http.Request) DataEncoding {
		return encoding
	}

	return func() {
		restoreParams()
		getDataEncoding = restoreGetDataEncoding
	}
}

func setBlocksFromParams(from, maxBlocks string) func() {
	restoreGetFrom := getFrom
	restoreGetMaxBlocks := getMaxBlocks

	getFrom = func(_ *http.Request) string {
		return from
	}

	getMaxBlocks = func(_ *http.Request) string {
		return maxBlocks
	}

	return func() {
		getFrom = restoreGetFrom
		getMaxBlocks = restoreGetMaxBlocks
	}
}

func setBlocksByHashParams(hash string) func() {
	restoreGetHash := getHash

	getHash = func(_ *http.Request) string {
		return hash
	}

	return func() {
		getHash = restoreGetHash
	}
}

func setBlocksByHashWithEncodingParams(hash string, encoding DataEncoding) func() {
	restoreParams := setBlocksByHashParams(hash)
	restoreGetDataEncoding := getDataEncoding

	getDataEncoding = func(_ *http.Request) DataEncoding {
		return encoding
	}

	return func() {
		restoreParams()
		getDataEncoding = restoreGetDataEncoding
	}
}
