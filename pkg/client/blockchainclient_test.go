/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package client

import (
	"testing"

	"github.com/hyperledger/fabric/protos/common"
	"github.com/stretchr/testify/require"
)

func TestBlockchainClientProvider(t *testing.T) {
	p := NewBlockchainProvider()
	require.NotNil(t, p)

	t.Run("Nil client", func(t *testing.T) {
		getLedger = func(channelID string) Blockchain { return nil }
		client, err := p.ForChannel(channel1)
		require.Error(t, err)
		require.Nil(t, client)
	})

	t.Run("Valid client", func(t *testing.T) {
		getLedger = func(channelID string) Blockchain { return &mockBlockchainClient{channelID: channelID} }
		client1, err := p.ForChannel(channel1)
		require.NoError(t, err)
		require.NotNil(t, client1)

		client2, err := p.ForChannel(channel2)
		require.NoError(t, err)
		require.NotNil(t, client2)
		require.NotEqual(t, client1, client2)
	})
}

type mockBlockchainClient struct {
	channelID string
}

func (c *mockBlockchainClient) GetBlockchainInfo() (*common.BlockchainInfo, error) {
	panic("not implemented")
}

func (c *mockBlockchainClient) GetBlockByNumber(blockNumber uint64) (*common.Block, error) {
	panic("not implemented")
}
