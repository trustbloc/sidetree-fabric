/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package store

import (
	"errors"
	"hash"
	"testing"

	"github.com/hyperledger/fabric-protos-go/ledger/queryresult"
	"github.com/stretchr/testify/require"
	"github.com/trustbloc/sidetree-core-go/pkg/api/operation"

	"github.com/trustbloc/sidetree-fabric/pkg/config"
	cfgmocks "github.com/trustbloc/sidetree-fabric/pkg/config/mocks"
	obmocks "github.com/trustbloc/sidetree-fabric/pkg/observer/mocks"
)

const (
	channel1 = "channel1"
	ns1      = "did:sidetree"
)

func TestProvider(t *testing.T) {
	vk1 := &queryresult.KV{
		Namespace: "document~diddoc",
		Key:       "did:sidetree:suffix",
		Value:     []byte("{}"),
	}

	offLedgerClient := obmocks.NewMockOffLedgerClient().
		WithDefaultQueryResults([]*queryresult.KV{vk1})

	cfgService := &cfgmocks.SidetreeConfigService{}

	t.Run("Get and Put", func(t *testing.T) {
		offLedgerProvider := &obmocks.OffLedgerClientProvider{}
		offLedgerProvider.ForChannelReturns(offLedgerClient, nil)

		p := NewProvider(channel1, cfgService, offLedgerProvider)
		require.NotNil(t, p)

		s, err := p.ForNamespace(ns1)
		require.NoError(t, err)
		require.NotNil(t, s)

		ops, err := s.Get("suffix")
		require.NoError(t, err)
		require.Len(t, ops, 1)

		require.NoError(t, s.Put([]*operation.AnchoredOperation{{Type: "create"}}))
	})

	t.Run("DCAS error", func(t *testing.T) {
		errExpected := errors.New("injected DCAS provider error")
		offLedgerProvider := &obmocks.OffLedgerClientProvider{}
		offLedgerProvider.ForChannelReturns(nil, errExpected)

		p := NewProvider(channel1, cfgService, offLedgerProvider)
		require.NotNil(t, p)

		s, err := p.ForNamespace(ns1)
		require.EqualError(t, err, errExpected.Error())
		require.Nil(t, s)
	})

	t.Run("Config service error", func(t *testing.T) {
		errExpected := errors.New("injected config service error")

		cfgService := &cfgmocks.SidetreeConfigService{}
		cfgService.LoadSidetreeReturns(config.Sidetree{}, errExpected)

		offLedgerProvider := &obmocks.OffLedgerClientProvider{}
		offLedgerProvider.ForChannelReturns(offLedgerClient, nil)

		p := NewProvider(channel1, cfgService, offLedgerProvider)
		require.NotNil(t, p)

		s, err := p.ForNamespace(ns1)
		require.EqualError(t, err, errExpected.Error())
		require.Nil(t, s)
	})
}

type mockHash struct {
	hash.Hash
	err error
}

func (m *mockHash) Write(p []byte) (n int, err error) {
	if m.err != nil {
		return -1, m.err
	}

	return m.Hash.Write(p)
}
