/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package blockchainhandler

import (
	"encoding/base64"
	"errors"
	"testing"

	cb "github.com/hyperledger/fabric-protos-go/common"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/protoutil"
	"github.com/stretchr/testify/require"
	"github.com/trustbloc/fabric-peer-ext/pkg/mocks"
	"github.com/trustbloc/sidetree-fabric/pkg/observer/common"
	obmocks "github.com/trustbloc/sidetree-fabric/pkg/observer/mocks"
)

func TestTxnValidator_Scan(t *testing.T) {
	const (
		blockNum1 = 1000
		blockNum2 = 1001

		txn1 = "tx1"
		txn2 = "tx2"

		anchor1 = "anchor1"
		anchor2 = "anchor2"
	)

	t.Run("Invalid descriptor", func(t *testing.T) {
		bcClient := &obmocks.BlockchainClient{}
		v := newFirstValidScanner(channel1, bcClient)
		require.NotNil(t, v)

		txns, foundValid, err := v.Scan(&mockDesc{}, 0)
		require.EqualError(t, err, errInvalidDesc.Error())
		require.Empty(t, txns)
		require.False(t, foundValid)
	})

	t.Run("Found valid", func(t *testing.T) {
		bb := mocks.NewBlockBuilder(channel1, blockNum1)
		bb.Transaction(txn1, pb.TxValidationCode_VALID).ChaincodeAction("sidetree").Write("some key", []byte("some value"))
		bb.Transaction(txn2, pb.TxValidationCode_VALID).ChaincodeAction("sidetree").Write(common.AnchorPrefix, getTxnInfo(anchor2))

		block := bb.Build()
		blockHash := base64.URLEncoding.EncodeToString(protoutil.BlockHeaderHash(block.Header))

		bcClient := &obmocks.BlockchainClient{}
		bcClient.GetBlockByNumberReturns(block, nil)

		desc := &firstValidDesc{Transaction{
			TransactionTime:     blockNum1,
			TransactionTimeHash: blockHash,
			TransactionNumber:   1,
			AnchorString:        anchor2,
		}}

		v := newFirstValidScanner(channel1, bcClient)
		require.NotNil(t, v)

		txns, foundValid, err := v.Scan(desc, 0)
		require.NoError(t, err)
		require.True(t, foundValid)
		require.Len(t, txns, 1)
		require.Equal(t, desc.Transaction(), txns[0])
	})

	t.Run("Invalid transaction info", func(t *testing.T) {
		bb := mocks.NewBlockBuilder(channel1, blockNum1)
		bb.Transaction(txn2, pb.TxValidationCode_VALID).ChaincodeAction("sidetree").Write(common.AnchorPrefix, []byte(anchor2))

		block := bb.Build()
		blockHash := base64.URLEncoding.EncodeToString(protoutil.BlockHeaderHash(block.Header))

		bcClient := &obmocks.BlockchainClient{}
		bcClient.GetBlockByNumberReturns(block, nil)

		desc := &firstValidDesc{Transaction{
			TransactionTime:     blockNum1,
			TransactionTimeHash: blockHash,
			TransactionNumber:   0,
			AnchorString:        anchor2,
		}}

		v := newFirstValidScanner(channel1, bcClient)
		require.NotNil(t, v)

		txns, foundValid, err := v.Scan(desc, 0)
		require.NoError(t, err)
		require.Empty(t, txns)
		require.False(t, foundValid)
	})

	t.Run("Invalid block number", func(t *testing.T) {
		bcClient := &obmocks.BlockchainClient{}
		bcClient.GetBlockByNumberReturns(nil, errors.New("not found"))

		desc := &firstValidDesc{Transaction{
			TransactionTime: blockNum2,
		}}

		v := newFirstValidScanner(channel1, bcClient)
		require.NotNil(t, v)

		txns, foundValid, err := v.Scan(desc, 0)
		require.NoError(t, err)
		require.Empty(t, txns)
		require.False(t, foundValid)
	})

	t.Run("Invalid base64 encoded block hash", func(t *testing.T) {
		bb := mocks.NewBlockBuilder(channel1, blockNum1)
		bb.Transaction(txn1, pb.TxValidationCode_VALID).ChaincodeAction("sidetree").Write(common.AnchorPrefix, getTxnInfo(anchor1))

		bcClient := &obmocks.BlockchainClient{}
		bcClient.GetBlockByNumberReturns(bb.Build(), nil)

		desc := &firstValidDesc{Transaction{
			TransactionTime:     blockNum2,
			TransactionTimeHash: "invalid base64",
		}}

		v := newFirstValidScanner(channel1, bcClient)
		require.NotNil(t, v)

		txns, foundValid, err := v.Scan(desc, 0)
		require.NoError(t, err)
		require.Empty(t, txns)
		require.False(t, foundValid)
	})

	t.Run("Invalid block hash", func(t *testing.T) {
		bb1 := mocks.NewBlockBuilder(channel1, blockNum1)
		bb1.Transaction(txn1, pb.TxValidationCode_VALID).ChaincodeAction("sidetree").Write(common.AnchorPrefix, getTxnInfo(anchor1))

		bb2 := mocks.NewBlockBuilder(channel1, blockNum2)
		bb2.Transaction(txn1, pb.TxValidationCode_VALID).ChaincodeAction("sidetree").Write(common.AnchorPrefix, getTxnInfo(anchor2))

		block1 := bb1.Build()
		block2 := bb2.Build()

		blockHash2 := base64.URLEncoding.EncodeToString(protoutil.BlockHeaderHash(block2.Header))

		bcClient := &obmocks.BlockchainClient{}
		bcClient.GetBlockByNumberReturns(block1, nil)

		desc := &firstValidDesc{Transaction{
			TransactionTime:     blockNum1,
			TransactionTimeHash: blockHash2,
		}}

		v := newFirstValidScanner(channel1, bcClient)
		require.NotNil(t, v)

		txns, foundValid, err := v.Scan(desc, 0)
		require.NoError(t, err)
		require.Empty(t, txns)
		require.False(t, foundValid)
	})

	t.Run("Invalid TransactionNumber", func(t *testing.T) {
		bb := mocks.NewBlockBuilder(channel1, blockNum1)
		bb.Transaction(txn1, pb.TxValidationCode_VALID).ChaincodeAction("sidetree").Write(common.AnchorPrefix, getTxnInfo(anchor1))

		block := bb.Build()
		blockHash := base64.URLEncoding.EncodeToString(protoutil.BlockHeaderHash(block.Header))

		bcClient := &obmocks.BlockchainClient{}
		bcClient.GetBlockByNumberReturns(block, nil)

		desc := &firstValidDesc{Transaction{
			TransactionNumber:   1,
			TransactionTime:     blockNum2,
			TransactionTimeHash: blockHash,
		}}

		v := newFirstValidScanner(channel1, bcClient)
		require.NotNil(t, v)

		txns, foundValid, err := v.Scan(desc, 0)
		require.NoError(t, err)
		require.Empty(t, txns)
		require.False(t, foundValid)
	})

	t.Run("Invalid anchor_string", func(t *testing.T) {
		bb := mocks.NewBlockBuilder(channel1, blockNum1)
		bb.Transaction(txn1, pb.TxValidationCode_VALID).ChaincodeAction("sidetree").Write(common.AnchorPrefix, getTxnInfo(anchor1))

		block := bb.Build()
		blockHash := base64.URLEncoding.EncodeToString(protoutil.BlockHeaderHash(block.Header))

		bcClient := &obmocks.BlockchainClient{}
		bcClient.GetBlockByNumberReturns(block, nil)

		desc := &firstValidDesc{Transaction{
			TransactionNumber:   0,
			TransactionTime:     blockNum1,
			TransactionTimeHash: blockHash,
			AnchorString:        "invalid anchor",
		}}

		v := newFirstValidScanner(channel1, bcClient)
		require.NotNil(t, v)

		txns, foundValid, err := v.Scan(desc, 0)
		require.NoError(t, err)
		require.Empty(t, txns)
		require.False(t, foundValid)
	})

	t.Run("Blockchain client error", func(t *testing.T) {
		errExpected := errors.New("injected blockchain client error")
		bcClient := &obmocks.BlockchainClient{}
		bcClient.GetBlockByNumberReturns(nil, errExpected)

		desc := &firstValidDesc{Transaction{
			TransactionNumber: 1,
			TransactionTime:   blockNum1,
		}}

		v := newFirstValidScanner(channel1, bcClient)
		require.NotNil(t, v)

		txns, foundValid, err := v.Scan(desc, 0)
		require.EqualError(t, err, errExpected.Error())
		require.Empty(t, txns)
		require.False(t, foundValid)
	})

	t.Run("Bad block -> ignore", func(t *testing.T) {
		block := &cb.Block{
			Header: &cb.BlockHeader{
				Number:       1000,
				PreviousHash: []byte("xxx"),
				DataHash:     []byte("yyy"),
			},
			Data: &cb.BlockData{
				Data: [][]byte{[]byte("invalid data")},
			},
		}

		blockHash := base64.URLEncoding.EncodeToString(protoutil.BlockHeaderHash(block.Header))

		bcClient := &obmocks.BlockchainClient{}
		bcClient.GetBlockByNumberReturns(block, nil)

		desc := &firstValidDesc{Transaction{
			TransactionNumber:   1,
			TransactionTimeHash: blockHash,
			TransactionTime:     blockNum1,
		}}

		v := newFirstValidScanner(channel1, bcClient)
		require.NotNil(t, v)

		txns, foundValid, err := v.Scan(desc, 0)
		require.NoError(t, err)
		require.Empty(t, txns)
		require.False(t, foundValid)
	})
}

type mockDesc struct {
	Transaction
}

func (d *mockDesc) BlockNum() uint64 {
	return d.TransactionTime
}

func (d *mockDesc) TxnNum() uint64 {
	return d.TransactionNumber
}

type event struct {
	id      int
	errChan chan error
}
