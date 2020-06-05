/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package blockchainhandler

import (
	"errors"
	"testing"

	cb "github.com/hyperledger/fabric-protos-go/common"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/stretchr/testify/require"
	"github.com/trustbloc/fabric-peer-ext/pkg/mocks"

	"github.com/trustbloc/sidetree-fabric/pkg/observer/common"
	obmocks "github.com/trustbloc/sidetree-fabric/pkg/observer/mocks"
)

func TestBlockchainScanner(t *testing.T) {
	const blockNum = 1000
	const txn1 = "tx1"
	const txn2 = "tx2"
	const anchor = "1.anchor"

	bb := mocks.NewBlockBuilder(channel1, blockNum)
	bb.Transaction(txn1, pb.TxValidationCode_VALID).ChaincodeAction("sidetree").Write(common.AnchorPrefix, getTxnInfo(anchor))
	bb.Transaction(txn2, pb.TxValidationCode_VALID).ChaincodeAction("sidetree").Write(common.AnchorPrefix, getTxnInfo(anchor))

	bcInfo := &cb.BlockchainInfo{
		Height: 2,
	}

	bcClient := &obmocks.BlockchainClient{}

	t.Run("Success", func(t *testing.T) {
		bcClient.GetBlockByNumberReturns(bb.Build(), nil)

		resp, err := scanBlockchain(channel1, 10, newBlockScanner(channel1, bcClient), newBlockIterator(bcInfo, 1, 0))
		require.NoError(t, err)
		require.Len(t, resp.Transactions, 2)
	})

	t.Run("Maximum reached", func(t *testing.T) {
		bcClient.GetBlockByNumberReturns(bb.Build(), nil)

		resp, err := scanBlockchain(channel1, 0, newBlockScanner(channel1, bcClient), newBlockIterator(bcInfo, 1, 0))
		require.NoError(t, err)
		require.True(t, resp.More)
		require.Empty(t, resp.Transactions)
	})

	t.Run("GetBlockByNumber error", func(t *testing.T) {
		errExpected := errors.New("injected blockchain client error")
		bcClient.GetBlockchainInfoReturns(bcInfo, nil)
		bcClient.GetBlockByNumberReturns(nil, errExpected)

		resp, err := scanBlockchain(channel1, 0, newBlockScanner(channel1, bcClient), newBlockIterator(bcInfo, 1, 0))
		require.Error(t, err)
		require.Contains(t, err.Error(), errExpected.Error())
		require.Nil(t, resp)
	})
}
