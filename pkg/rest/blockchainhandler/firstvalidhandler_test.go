/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package blockchainhandler

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	cb "github.com/hyperledger/fabric-protos-go/common"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/protoutil"
	"github.com/stretchr/testify/require"
	"github.com/trustbloc/fabric-peer-ext/pkg/mocks"

	"github.com/trustbloc/sidetree-fabric/pkg/httpserver"
	"github.com/trustbloc/sidetree-fabric/pkg/observer/common"
	obmocks "github.com/trustbloc/sidetree-fabric/pkg/observer/mocks"
)

func TestNewFirstValidHandler(t *testing.T) {
	bcProvider := &obmocks.BlockchainClientProvider{}

	h := NewFirstValidHandler(channel1, handlerCfg, bcProvider)
	require.NotNil(t, h)

	require.Equal(t, "/blockchain/firstValid", h.Path())
	require.Equal(t, http.MethodPost, h.Method())
}

func TestFirstValid_Handler(t *testing.T) {
	const blockNum = 1000
	const txn1 = "tx1"
	const txn2 = "tx2"
	const anchor1 = "anchor1"
	const anchor2 = "anchor2"
	const anchor3 = "anchor3"

	bcInfo := &cb.BlockchainInfo{
		Height:           1001,
		CurrentBlockHash: []byte{1, 2, 3, 4},
	}

	bcProvider := &obmocks.BlockchainClientProvider{}

	t.Run("Success", func(t *testing.T) {
		bb := mocks.NewBlockBuilder(channel1, blockNum)
		bb.Transaction(txn1, pb.TxValidationCode_VALID).ChaincodeAction("sidetree").Write(common.AnchorAddrPrefix, []byte(anchor1))
		bb.Transaction(txn2, pb.TxValidationCode_VALID).ChaincodeAction("sidetree").Write(common.AnchorAddrPrefix, []byte(anchor2))
		bb.Transaction(txn2, pb.TxValidationCode_VALID).ChaincodeAction("sidetree").Write(common.AnchorAddrPrefix, []byte(anchor3))

		block := bb.Build()
		blockHash := base64.URLEncoding.EncodeToString(protoutil.BlockHeaderHash(block.Header))

		bcClient := &obmocks.BlockchainClient{}
		bcClient.GetBlockchainInfoReturns(bcInfo, nil)
		bcClient.GetBlockByNumberReturns(block, nil)
		bcProvider.ForChannelReturns(bcClient, nil)

		h := NewFirstValidHandler(channel1, handlerCfg, bcProvider)
		require.NotNil(t, h)

		txns := []Transaction{
			{
				TransactionTime:     blockNum,
				TransactionTimeHash: blockHash,
				TransactionNumber:   0,
				AnchorString:        "invalid anchor",
			},
			{
				TransactionTime:     blockNum,
				TransactionTimeHash: blockHash,
				TransactionNumber:   1,
				AnchorString:        anchor2,
			},
			{
				TransactionTime:     blockNum,
				TransactionTimeHash: blockHash,
				TransactionNumber:   2,
				AnchorString:        anchor3,
			},
		}

		txnsBytes, err := json.Marshal(txns)
		require.NoError(t, err)

		rw := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/blockchain/firstValid", bytes.NewReader(txnsBytes))

		h.Handler()(rw, req)

		require.Equal(t, http.StatusOK, rw.Result().StatusCode)
		require.Equal(t, httpserver.ContentTypeJSON, rw.Header().Get(httpserver.ContentTypeHeader))

		resp := Transaction{}
		require.NoError(t, json.Unmarshal(rw.Body.Bytes(), &resp))
		require.Equal(t, txns[1], resp)
	})

	t.Run("No valid transactions", func(t *testing.T) {
		bb := mocks.NewBlockBuilder(channel1, blockNum)
		bb.Transaction(txn1, pb.TxValidationCode_VALID).ChaincodeAction("sidetree").Write(common.AnchorAddrPrefix, []byte(anchor1))

		block := bb.Build()
		blockHash := base64.URLEncoding.EncodeToString(protoutil.BlockHeaderHash(block.Header))

		bcClient := &obmocks.BlockchainClient{}
		bcClient.GetBlockchainInfoReturns(bcInfo, nil)
		bcClient.GetBlockByNumberReturns(block, nil)
		bcProvider.ForChannelReturns(bcClient, nil)

		h := NewFirstValidHandler(channel1, handlerCfg, bcProvider)
		require.NotNil(t, h)

		txns := []Transaction{
			{
				TransactionTime:     blockNum,
				TransactionTimeHash: blockHash,
				TransactionNumber:   0,
				AnchorString:        "invalid anchor",
			},
		}

		txnsBytes, err := json.Marshal(txns)
		require.NoError(t, err)

		rw := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/blockchain/firstValid", bytes.NewReader(txnsBytes))

		h.Handler()(rw, req)

		require.Equal(t, http.StatusNotFound, rw.Result().StatusCode)
		require.Equal(t, httpserver.ContentTypeText, rw.Header().Get(httpserver.ContentTypeHeader))
		require.Equal(t, httpserver.StatusNotFound, rw.Body.String())
	})

	t.Run("Blockchain provider error", func(t *testing.T) {
		errExpected := errors.New("injected blockchain provider error")
		bcProvider.ForChannelReturns(nil, errExpected)

		h := NewFirstValidHandler(channel1, handlerCfg, bcProvider)
		require.NotNil(t, h)

		txnsBytes, err := json.Marshal([]Transaction{})
		require.NoError(t, err)

		rw := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/blockchain/firstValid", bytes.NewReader(txnsBytes))

		h.Handler()(rw, req)

		require.Equal(t, http.StatusInternalServerError, rw.Result().StatusCode)
		require.Equal(t, httpserver.ContentTypeText, rw.Header().Get(httpserver.ContentTypeHeader))
		require.Equal(t, httpserver.StatusServerError, rw.Body.String())
	})

	t.Run("Invalid post data", func(t *testing.T) {
		bcClient := &obmocks.BlockchainClient{}
		bcProvider.ForChannelReturns(bcClient, nil)

		h := NewFirstValidHandler(channel1, handlerCfg, bcProvider)
		require.NotNil(t, h)

		rw := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/blockchain/firstValid", bytes.NewReader([]byte("invalid data")))

		h.Handler()(rw, req)

		require.Equal(t, http.StatusBadRequest, rw.Result().StatusCode)
		require.Equal(t, httpserver.ContentTypeText, rw.Header().Get(httpserver.ContentTypeHeader))
		require.Equal(t, httpserver.StatusBadRequest, rw.Body.String())
	})

	t.Run("Marshal error", func(t *testing.T) {
		bb := mocks.NewBlockBuilder(channel1, blockNum)
		bb.Transaction(txn1, pb.TxValidationCode_VALID).ChaincodeAction("sidetree").Write(common.AnchorAddrPrefix, []byte(anchor1))

		block := bb.Build()
		blockHash := base64.URLEncoding.EncodeToString(protoutil.BlockHeaderHash(block.Header))

		bcClient := &obmocks.BlockchainClient{}
		bcClient.GetBlockchainInfoReturns(bcInfo, nil)
		bcClient.GetBlockByNumberReturns(block, nil)
		bcProvider.ForChannelReturns(bcClient, nil)

		h := NewFirstValidHandler(channel1, handlerCfg, bcProvider)
		require.NotNil(t, h)
		h.jsonMarshal = func(v interface{}) ([]byte, error) { return nil, errors.New("injected marshal error") }

		txns := []Transaction{
			{
				TransactionTime:     blockNum,
				TransactionTimeHash: blockHash,
				TransactionNumber:   0,
				AnchorString:        anchor1,
			},
		}

		txnsBytes, err := json.Marshal(txns)
		require.NoError(t, err)

		rw := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/blockchain/firstValid", bytes.NewReader(txnsBytes))

		h.Handler()(rw, req)

		require.Equal(t, http.StatusInternalServerError, rw.Result().StatusCode)
		require.Equal(t, httpserver.ContentTypeText, rw.Header().Get(httpserver.ContentTypeHeader))
		require.Equal(t, httpserver.StatusServerError, rw.Body.String())
	})
}
