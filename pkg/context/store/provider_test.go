/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package store

import (
	"errors"
	"testing"

	"github.com/hyperledger/fabric-protos-go/ledger/queryresult"
	"github.com/stretchr/testify/require"
	extmocks "github.com/trustbloc/fabric-peer-ext/pkg/mocks"
	"github.com/trustbloc/sidetree-core-go/pkg/api/batch"

	"github.com/trustbloc/sidetree-fabric/pkg/config"
	cfgmocks "github.com/trustbloc/sidetree-fabric/pkg/config/mocks"
	"github.com/trustbloc/sidetree-fabric/pkg/mocks"
)

const (
	channel1 = "channel1"
	ns1      = "did:sidetree"
)

func TestProvider(t *testing.T) {
	dcasClient := &mocks.DCASClient{}

	vk1 := &queryresult.KV{
		Namespace: "document~diddoc",
		Key:       "did:sidetree:suffix",
		Value:     []byte("{}"),
	}
	it := extmocks.NewResultsIterator().WithResults([]*queryresult.KV{vk1})
	dcasClient.QueryReturns(it, nil)

	cfgService := &cfgmocks.SidetreeConfigService{}

	t.Run("Get and Put", func(t *testing.T) {
		dcasProvider := &mocks.DCASClientProvider{}
		dcasProvider.ForChannelReturns(dcasClient, nil)

		p := NewProvider(channel1, cfgService, dcasProvider)
		require.NotNil(t, p)

		s, err := p.ForNamespace(ns1)
		require.NoError(t, err)
		require.NotNil(t, s)

		ops, err := s.Get("suffix")
		require.NoError(t, err)
		require.Len(t, ops, 1)

		require.NoError(t, s.Put([]*batch.AnchoredOperation{{Type: "create"}}))
	})

	t.Run("DCAS error", func(t *testing.T) {
		errExpected := errors.New("injected DCAS provider error")
		dcasProvider := &mocks.DCASClientProvider{}
		dcasProvider.ForChannelReturns(nil, errExpected)

		p := NewProvider(channel1, cfgService, dcasProvider)
		require.NotNil(t, p)

		s, err := p.ForNamespace(ns1)
		require.EqualError(t, err, errExpected.Error())
		require.Nil(t, s)
	})

	t.Run("Config service error", func(t *testing.T) {
		errExpected := errors.New("injected config service error")

		cfgService := &cfgmocks.SidetreeConfigService{}
		cfgService.LoadSidetreeReturns(config.Sidetree{}, errExpected)

		dcasProvider := &mocks.DCASClientProvider{}
		dcasProvider.ForChannelReturns(dcasClient, nil)

		p := NewProvider(channel1, cfgService, dcasProvider)
		require.NotNil(t, p)

		s, err := p.ForNamespace(ns1)
		require.EqualError(t, err, errExpected.Error())
		require.Nil(t, s)
	})
}
