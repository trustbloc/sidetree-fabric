/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package client

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOffLedgerClientProvider(t *testing.T) {
	p := NewOffLedgerProvider()
	require.NotNil(t, p)

	client1 := p.ForChannel(channel1)
	require.NotNil(t, client1)

	client2 := p.ForChannel(channel2)
	require.NotNil(t, client2)
	require.NotEqual(t, client1, client2)
}
