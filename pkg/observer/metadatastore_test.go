/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package observer

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/trustbloc/fabric-peer-ext/pkg/roles"

	obmocks "github.com/trustbloc/sidetree-fabric/pkg/observer/mocks"
	"github.com/trustbloc/sidetree-fabric/pkg/peer/mocks"
)

const (
	cc1 = "cc1"
)

// Initialize roles
var _ = roles.GetRoles()

func TestMetaDataStore(t *testing.T) {
	olp := &obmocks.OffLedgerClientProvider{}
	peerCfg := &mocks.PeerConfig{}
	peerCfg.MSPIDReturns(org1)
	peerCfg.PeerIDReturns(peer1)

	t.Run("Get - provider error", func(t *testing.T) {
		s := NewMetaDataStore(channel1, peerCfg, cc1, olp)
		require.NotNil(t, s)

		errProvider := errors.New("injected provider error")
		olp.ForChannelReturns(nil, errProvider)

		d, err := s.Get()
		require.EqualError(t, err, errProvider.Error())
		require.Nil(t, d)
	})

	t.Run("Put - provider error", func(t *testing.T) {
		s := NewMetaDataStore(channel1, peerCfg, cc1, olp)
		require.NotNil(t, s)

		errProvider := errors.New("injected provider error")
		olp.ForChannelReturns(nil, errProvider)

		err := s.Put(&Metadata{})
		require.EqualError(t, err, errProvider.Error())
	})

	t.Run("Get - client error", func(t *testing.T) {
		s := NewMetaDataStore(channel1, peerCfg, cc1, olp)
		require.NotNil(t, s)

		ols := obmocks.NewMockOffLedgerClient()
		olp.ForChannelReturns(ols, nil)

		ols.GetErr = errors.New("injected Get error")

		d, err := s.Get()
		require.Error(t, err)
		require.Contains(t, err.Error(), ols.GetErr.Error())
		require.Nil(t, d)
	})

	t.Run("Put - client error", func(t *testing.T) {
		s := NewMetaDataStore(channel1, peerCfg, cc1, olp)
		require.NotNil(t, s)

		ols := obmocks.NewMockOffLedgerClient()
		olp.ForChannelReturns(ols, nil)

		ols.PutErr = errors.New("injected Put error")

		err := s.Put(&Metadata{})
		require.Error(t, err)
		require.Contains(t, err.Error(), ols.PutErr.Error())
	})

	t.Run("Not found error", func(t *testing.T) {
		s := NewMetaDataStore(channel1, peerCfg, cc1, olp)
		require.NotNil(t, s)

		ols := obmocks.NewMockOffLedgerClient()
		olp.ForChannelReturns(ols, nil)

		d, err := s.Get()
		require.EqualError(t, err, errMetaDataNotFound.Error())
		require.Nil(t, d)
	})

	t.Run("Get and put -> success", func(t *testing.T) {
		s := NewMetaDataStore(channel1, peerCfg, cc1, olp)
		require.NotNil(t, s)

		ols := obmocks.NewMockOffLedgerClient()
		olp.ForChannelReturns(ols, nil)

		data := &Metadata{}
		require.NoError(t, s.Put(data))

		d, err := s.Get()
		require.NoError(t, err)
		require.Equal(t, data, d)
	})

	t.Run("Clustered -> success", func(t *testing.T) {
		rolesValue := make(map[roles.Role]struct{})
		rolesValue[roles.CommitterRole] = struct{}{}
		roles.SetRoles(rolesValue)

		defer func() {
			roles.SetRoles(nil)
		}()

		s := NewMetaDataStore(channel1, peerCfg, cc1, olp)
		require.NotNil(t, s)

		ols := obmocks.NewMockOffLedgerClient()
		olp.ForChannelReturns(ols, nil)

		data := &Metadata{}
		require.NoError(t, s.Put(data))

		d, err := s.Get()
		require.NoError(t, err)
		require.Equal(t, data, d)
	})
}
