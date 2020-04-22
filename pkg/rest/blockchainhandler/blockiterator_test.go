/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package blockchainhandler

import (
	"testing"

	cb "github.com/hyperledger/fabric-protos-go/common"
	"github.com/stretchr/testify/require"
)

func TestSinceTimeIterator_Next(t *testing.T) {
	const blockNum1 = uint64(1000)
	const blockNum2 = uint64(1001)
	const txnNum = uint64(1)

	it := newBlockIterator(&cb.BlockchainInfo{Height: 1002}, blockNum1, txnNum)
	require.NotNil(t, it)

	desc, ok := it.Next()
	require.True(t, ok)
	require.NotNil(t, desc)
	require.Equal(t, blockNum1, desc.BlockNum())
	require.Equal(t, txnNum, desc.TxnNum())

	desc, ok = it.Next()
	require.True(t, ok)
	require.NotNil(t, desc)
	require.Equal(t, blockNum2, desc.BlockNum())
	require.Equal(t, uint64(0), desc.TxnNum())

	desc, ok = it.Next()
	require.False(t, ok)
	require.Nil(t, desc)
}
