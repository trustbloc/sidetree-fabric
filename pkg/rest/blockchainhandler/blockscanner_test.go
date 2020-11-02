/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package blockchainhandler

import (
	"fmt"
	"testing"

	cb "github.com/hyperledger/fabric-protos-go/common"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/stretchr/testify/require"
	"github.com/trustbloc/fabric-peer-ext/pkg/mocks"

	"github.com/trustbloc/sidetree-fabric/pkg/observer/common"
	obmocks "github.com/trustbloc/sidetree-fabric/pkg/observer/mocks"
)

func TestBlockScanner(t *testing.T) {
	const blockNum = 1000
	const maxTxns = 2

	const txn1 = "tx1"
	const txn2 = "tx2"
	const txn3 = "tx3"

	const anchor1 = "anchor1"
	const anchor2 = "anchor2"
	const anchor3 = "anchor3"

	t.Run("Success", func(t *testing.T) {
		bb := mocks.NewBlockBuilder(channel1, blockNum)
		bb.Transaction(txn1, pb.TxValidationCode_VALID).ChaincodeAction("sidetree").Write(common.AnchorPrefix, getTxnInfo(anchor1))
		bb.Transaction(txn1, pb.TxValidationCode_VALID).ChaincodeAction("sidetree").Write(common.AnchorPrefix, getTxnInfo(anchor2))

		bcClient := &obmocks.BlockchainClient{}
		bcClient.GetBlockByNumberReturns(bb.Build(), nil)

		scanner := newBlockScanner(channel1, bcClient)
		require.NotNil(t, scanner)

		desc := &blockDesc{
			blockNum: blockNum,
			txnNum:   1,
		}

		txns, reachedMax, err := scanner.Scan(desc, maxTxns)
		require.NoError(t, err)
		require.Len(t, txns, 1)
		require.Equal(t, anchor2, txns[0].AnchorString)
		require.False(t, reachedMax)
	})

	t.Run("Error - unmarshall transaction info", func(t *testing.T) {
		bb := mocks.NewBlockBuilder(channel1, blockNum)
		bb.Transaction(txn1, pb.TxValidationCode_VALID).ChaincodeAction("sidetree").Write(common.AnchorPrefix, []byte(anchor1))

		bcClient := &obmocks.BlockchainClient{}
		bcClient.GetBlockByNumberReturns(bb.Build(), nil)

		scanner := newBlockScanner(channel1, bcClient)
		require.NotNil(t, scanner)

		desc := &blockDesc{
			blockNum: blockNum,
			txnNum:   0,
		}

		txns, reachedMax, err := scanner.Scan(desc, maxTxns)
		require.NoError(t, err)
		require.Len(t, txns, 0)
		require.False(t, reachedMax)
	})

	t.Run("Maximum reached", func(t *testing.T) {
		bb := mocks.NewBlockBuilder(channel1, blockNum)
		bb.Transaction(txn1, pb.TxValidationCode_VALID).ChaincodeAction("sidetree").Write(common.AnchorPrefix, getTxnInfo(anchor1))
		bb.Transaction(txn2, pb.TxValidationCode_VALID).ChaincodeAction("sidetree").Write(common.AnchorPrefix, getTxnInfo(anchor2))
		bb.Transaction(txn3, pb.TxValidationCode_VALID).ChaincodeAction("sidetree").Write(common.AnchorPrefix, getTxnInfo(anchor3))

		bcClient := &obmocks.BlockchainClient{}
		bcClient.GetBlockByNumberReturns(bb.Build(), nil)

		scanner := newBlockScanner(channel1, bcClient)
		require.NotNil(t, scanner)

		desc := &blockDesc{
			blockNum: blockNum,
			txnNum:   0,
		}

		txns, reachedMax, err := scanner.Scan(desc, maxTxns)
		require.NoError(t, err)
		require.Len(t, txns, maxTxns)
		require.True(t, reachedMax)
	})

	t.Run("Irrelevant transaction", func(t *testing.T) {
		bb := mocks.NewBlockBuilder(channel1, blockNum)
		bb.Transaction(txn1, pb.TxValidationCode_VALID).ChaincodeAction("sidetree").Write("xxx", getTxnInfo("xxx"))

		bcClient := &obmocks.BlockchainClient{}
		bcClient.GetBlockByNumberReturns(bb.Build(), nil)

		scanner := newBlockScanner(channel1, bcClient)
		require.NotNil(t, scanner)

		desc := &blockDesc{
			blockNum: blockNum,
			txnNum:   1,
		}

		txns, reachedMax, err := scanner.Scan(desc, maxTxns)
		require.NoError(t, err)
		require.Empty(t, txns)
		require.False(t, reachedMax)
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

		bcClient := &obmocks.BlockchainClient{}
		bcClient.GetBlockByNumberReturns(block, nil)

		scanner := newBlockScanner(channel1, bcClient)
		require.NotNil(t, scanner)

		desc := &blockDesc{
			blockNum: blockNum,
			txnNum:   1,
		}

		txns, reachedMax, err := scanner.Scan(desc, maxTxns)
		require.NoError(t, err)
		require.Empty(t, txns)
		require.False(t, reachedMax)
	})
}

func getTxnInfo(anchor string) []byte {
	return []byte(fmt.Sprintf(`{"anchorString":"%s", "namespace": "ns"}`, anchor))
}
