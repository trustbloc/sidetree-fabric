/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package client

import (
	"testing"

	"github.com/hyperledger/fabric/protos/common"
	"github.com/stretchr/testify/require"
	"github.com/trustbloc/fabric-peer-ext/pkg/mocks"
)

const (
	channel1 = "channel1"
	channel2 = "channel2"
)

func TestBlockchainClientProvider(t *testing.T) {

	t.Run("Nil client", func(t *testing.T) {
		ledgerProvider := &mocks.LedgerProvider{}
		p := NewBlockchainProvider(ledgerProvider)
		require.NotNil(t, p)

		client, err := p.ForChannel(channel1)
		require.Error(t, err)
		require.Nil(t, client)
	})

	t.Run("Valid client", func(t *testing.T) {
		ledgerProvider := &mocks.LedgerProvider{}
		ledgerProvider.GetLedgerReturnsOnCall(0, &mocks.Ledger{BlockchainInfo: &common.BlockchainInfo{Height: 1000}})
		ledgerProvider.GetLedgerReturnsOnCall(1, &mocks.Ledger{BlockchainInfo: &common.BlockchainInfo{Height: 2000}})

		p := NewBlockchainProvider(ledgerProvider)
		require.NotNil(t, p)

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
