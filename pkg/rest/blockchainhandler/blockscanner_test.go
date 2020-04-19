/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package blockchainhandler

import (
	"testing"

	cb "github.com/hyperledger/fabric-protos-go/common"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/stretchr/testify/require"
	"github.com/trustbloc/fabric-peer-ext/pkg/mocks"
	"github.com/trustbloc/sidetree-fabric/pkg/observer/common"
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
		bb.Transaction(txn1, pb.TxValidationCode_VALID).ChaincodeAction("sidetree").Write(common.AnchorAddrPrefix, []byte(anchor1))

		p := newBlockScanner(channel1, bb.Build(), maxTxns)
		require.NotNil(t, p)

		txns, err := p.scan()
		require.NoError(t, err)
		require.Len(t, txns, 1)
		require.Equal(t, anchor1, txns[0].AnchorString)
	})

	t.Run("Maximum reached", func(t *testing.T) {
		bb := mocks.NewBlockBuilder(channel1, blockNum)
		bb.Transaction(txn1, pb.TxValidationCode_VALID).ChaincodeAction("sidetree").Write(common.AnchorAddrPrefix, []byte(anchor1))
		bb.Transaction(txn2, pb.TxValidationCode_VALID).ChaincodeAction("sidetree").Write(common.AnchorAddrPrefix, []byte(anchor2))
		bb.Transaction(txn3, pb.TxValidationCode_VALID).ChaincodeAction("sidetree").Write(common.AnchorAddrPrefix, []byte(anchor3))

		p := newBlockScanner(channel1, bb.Build(), maxTxns)
		require.NotNil(t, p)

		txns, err := p.scan()
		require.Error(t, err, errReachedMaxTxns.Error())
		require.Len(t, txns, maxTxns)
	})

	t.Run("Irrelevant transaction", func(t *testing.T) {
		bb := mocks.NewBlockBuilder(channel1, blockNum)
		bb.Transaction(txn1, pb.TxValidationCode_VALID).ChaincodeAction("sidetree").Write("xxx", []byte("xxx"))

		p := newBlockScanner(channel1, bb.Build(), maxTxns)
		require.NotNil(t, p)

		txns, err := p.scan()
		require.NoError(t, err)
		require.Empty(t, txns)
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

		p := newBlockScanner(channel1, block, maxTxns)
		require.NotNil(t, p)

		txns, err := p.scan()
		require.NoError(t, err)
		require.Empty(t, txns)
	})
}
