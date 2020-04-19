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
	const anchor = "xxx"

	bb := mocks.NewBlockBuilder(channel1, blockNum)
	bb.Transaction(txn1, pb.TxValidationCode_VALID).ChaincodeAction("sidetree").Write(common.AnchorAddrPrefix, []byte(anchor))
	bb.Transaction(txn2, pb.TxValidationCode_VALID).ChaincodeAction("sidetree").Write(common.AnchorAddrPrefix, []byte(anchor))

	bcInfo := &cb.BlockchainInfo{
		Height: 2,
	}

	bcClient := &obmocks.BlockchainClient{}

	t.Run("Success", func(t *testing.T) {
		bcClient.GetBlockchainInfoReturns(bcInfo, nil)
		bcClient.GetBlockByNumberReturns(bb.Build(), nil)

		v := newBlockchainScanner(channel1, 1, bcClient)
		require.NotNil(t, v)
		resp, err := v.scan()
		require.NoError(t, err)
		require.Len(t, resp.Transactions, 1)
	})

	t.Run("Maximum reached", func(t *testing.T) {
		bcClient.GetBlockchainInfoReturns(bcInfo, nil)
		bcClient.GetBlockByNumberReturns(bb.Build(), nil)

		v := newBlockchainScanner(channel1, 1, bcClient)
		require.NotNil(t, v)
		resp, err := v.scan()
		require.NoError(t, err)
		require.True(t, resp.More)
		require.Len(t, resp.Transactions, 1)
	})

	t.Run("Blockchain client error", func(t *testing.T) {
		errExpected := errors.New("injected blockchain client error")
		bcClient.GetBlockchainInfoReturns(nil, errExpected)

		v := newBlockchainScanner(channel1, 1, bcClient)
		require.NotNil(t, v)
		resp, err := v.scan()
		require.Error(t, err)
		require.Contains(t, err.Error(), errExpected.Error())
		require.Nil(t, resp)
	})

	t.Run("GetBlockByNumber error", func(t *testing.T) {
		errExpected := errors.New("injected blockchain client error")
		bcClient.GetBlockchainInfoReturns(bcInfo, nil)
		bcClient.GetBlockByNumberReturns(nil, errExpected)

		v := newBlockchainScanner(channel1, 1, bcClient)
		require.NotNil(t, v)
		resp, err := v.scan()
		require.Error(t, err)
		require.Contains(t, err.Error(), errExpected.Error())
		require.Nil(t, resp)
	})
}
