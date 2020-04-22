/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package blockchainhandler

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFirstValidIterator_Next(t *testing.T) {
	txns := []Transaction{
		{
			TransactionTime:   1000,
			TransactionNumber: 0,
		},
		{
			TransactionTime:   1001,
			TransactionNumber: 0,
		},
	}

	it := newFirstValidIterator(txns)
	require.NotNil(t, it)

	desc, ok := it.Next()
	require.True(t, ok)

	fvDesc, ok := desc.(*firstValidDesc)
	require.True(t, ok)
	require.Equal(t, txns[0], fvDesc.Transaction())

	desc, ok = it.Next()
	require.True(t, ok)

	fvDesc, ok = desc.(*firstValidDesc)
	require.True(t, ok)
	require.Equal(t, txns[1], fvDesc.Transaction())

	desc, ok = it.Next()
	require.False(t, ok)
	require.Nil(t, desc)
}
