/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package operationfilter

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/trustbloc/sidetree-fabric/pkg/mocks"

	coremocks "github.com/trustbloc/sidetree-core-go/pkg/mocks"
)

const (
	channel1 = "channel1"
	ns1      = "namespace1"
	ns2      = "namespace2"
)

func TestProvider(t *testing.T) {
	pcp := coremocks.NewMockProtocolClientProvider()
	pcp.ProtocolClients[ns1] = coremocks.NewMockProtocolClient()
	pcp.ProtocolClients[ns2] = coremocks.NewMockProtocolClient()

	opStoreProvider := &mocks.OperationStoreProvider{}

	p := NewProvider(channel1, opStoreProvider, pcp)
	require.NotNil(t, p)

	s1, err := p.Get(ns1)
	require.NoError(t, err)
	require.NotNil(t, s1)

	s2, err := p.Get(ns2)
	require.NoError(t, err)
	require.NotNil(t, s2)
	require.NotEqual(t, s1, s2)

	s3, err := p.Get(ns2)
	require.Equal(t, s2, s3)
}

func TestStoreProviderError(t *testing.T) {
	errExpected := errors.New("injected op store provider error")
	opStoreProvider := &mocks.OperationStoreProvider{}
	opStoreProvider.ForNamespaceReturns(nil, errExpected)

	pcp := coremocks.NewMockProtocolClientProvider()

	p := NewProvider(channel1, opStoreProvider, pcp)
	require.NotNil(t, p)

	s, err := p.Get(ns1)
	require.EqualError(t, err, errExpected.Error())
	require.Nil(t, s)
}

func TestProtocolProviderError(t *testing.T) {
	opStoreProvider := &mocks.OperationStoreProvider{}
	pcp := coremocks.NewMockProtocolClientProvider()

	p := NewProvider(channel1, opStoreProvider, pcp)
	require.NotNil(t, p)

	s, err := p.Get(ns1)
	require.Contains(t, err.Error(), "protocol client not found for namespace")
	require.Nil(t, s)
}
