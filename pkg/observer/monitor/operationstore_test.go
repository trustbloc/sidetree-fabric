/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package monitor

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/trustbloc/sidetree-core-go/pkg/api/batch"

	"github.com/trustbloc/sidetree-fabric/pkg/mocks"
)

const (
	ns1 = "ns1"
)

func TestOperationStore_Put(t *testing.T) {
	op1 := &batch.Operation{ID: "op1"}
	op2 := &batch.Operation{ID: "op2"}

	opStore := &mocks.OperationStore{}
	opStore.GetReturnsOnCall(0, []*batch.Operation{op1}, nil)

	s := NewOperationStore(channel1, opStore)
	require.NotNil(t, s)

	require.NoError(t, s.Put([]*batch.Operation{op1, op2}))
	require.Equal(t, 1, opStore.PutCallCount())
	require.Equal(t, []*batch.Operation{op2}, opStore.Invocations()["Put"][0][0])
}

func TestOperationStore_PutError(t *testing.T) {
	t.Run("DCAS get error", func(t *testing.T) {
		opStore := &mocks.OperationStore{}
		opStore.GetReturns(nil, errors.New("injected Get error"))

		s := NewOperationStore(channel1, opStore)

		err := s.Put([]*batch.Operation{{ID: "op1"}})
		require.Error(t, err)

		merr, ok := err.(monitorError)
		require.True(t, ok)
		require.True(t, merr.Transient())
	})

	t.Run("DCAS put error", func(t *testing.T) {
		opStore := &mocks.OperationStore{}
		opStore.PutReturns(errors.New("injected Put error"))

		s := NewOperationStore(channel1, opStore)
		err := s.Put([]*batch.Operation{{ID: "op1"}})
		require.Error(t, err)

		merr, ok := err.(monitorError)
		require.True(t, ok)
		require.True(t, merr.Transient())
	})
}

func TestOperationStoreProvider(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		provider := NewOperationStoreProvider(channel1, &mocks.OperationStoreProvider{})
		require.NotNil(t, provider)

		s, err := provider.ForNamespace(ns1)
		require.NoError(t, err)
		require.NotNil(t, s)
	})

	t.Run("Provider error", func(t *testing.T) {
		errExpected := errors.New("injected provider error")

		p := &mocks.OperationStoreProvider{}
		p.ForNamespaceReturns(nil, errExpected)

		provider := NewOperationStoreProvider(channel1, p)
		require.NotNil(t, provider)

		s, err := provider.ForNamespace(ns1)
		require.Error(t, err)
		require.Contains(t, err.Error(), errExpected.Error())
		require.Nil(t, s)
	})
}
