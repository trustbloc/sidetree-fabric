/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package monitor

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/trustbloc/sidetree-fabric/pkg/observer/mocks"
)

const (
	cc1 = "cc1"
)

func TestMetaDataStore(t *testing.T) {
	olp := &mocks.OffLedgerClientProvider{}

	s := NewMetaDataStore(channel1, peer1, cc1, olp)
	require.NotNil(t, s)

	t.Run("Get - provider error", func(t *testing.T) {
		errProvider := errors.New("injected provider error")
		olp.ForChannelReturns(nil, errProvider)

		d, err := s.Get()
		require.EqualError(t, err, errProvider.Error())
		require.Nil(t, d)
	})

	t.Run("Put - provider error", func(t *testing.T) {
		errProvider := errors.New("injected provider error")
		olp.ForChannelReturns(nil, errProvider)

		err := s.Put(&MetaData{})
		require.EqualError(t, err, errProvider.Error())
	})

	t.Run("Get - client error", func(t *testing.T) {
		ols := mocks.NewMockOffLedgerClient()
		olp.ForChannelReturns(ols, nil)

		ols.GetErr = errors.New("injected Get error")

		d, err := s.Get()
		require.Error(t, err)
		require.Contains(t, err.Error(), ols.GetErr.Error())
		require.Nil(t, d)
	})

	t.Run("Put - client error", func(t *testing.T) {
		ols := mocks.NewMockOffLedgerClient()
		olp.ForChannelReturns(ols, nil)

		ols.PutErr = errors.New("injected Put error")

		err := s.Put(&MetaData{})
		require.Error(t, err)
		require.Contains(t, err.Error(), ols.PutErr.Error())
	})

	t.Run("Not found error", func(t *testing.T) {
		ols := mocks.NewMockOffLedgerClient()
		olp.ForChannelReturns(ols, nil)

		d, err := s.Get()
		require.EqualError(t, err, errMetaDataNotFound.Error())
		require.Nil(t, d)
	})

	t.Run("Get and put -> success", func(t *testing.T) {
		ols := mocks.NewMockOffLedgerClient()
		olp.ForChannelReturns(ols, nil)

		data := &MetaData{}
		require.NoError(t, s.Put(data))

		d, err := s.Get()
		require.NoError(t, err)
		require.Equal(t, data, d)
	})
}
