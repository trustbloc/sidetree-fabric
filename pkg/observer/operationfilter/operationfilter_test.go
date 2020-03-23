/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package operationfilter

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/trustbloc/sidetree-fabric/pkg/observer/mocks"
)

const (
	channel1 = "channel1"
	ns1      = "namespace1"
	ns2      = "namespace2"
)

func TestProvider(t *testing.T) {
	opStoreProvider := &mocks.OpStoreClientProvider{}

	p := NewProvider(channel1, opStoreProvider)
	require.NotNil(t, p)

	s1 := p.Get(ns1)
	require.NotNil(t, s1)

	s2 := p.Get(ns2)
	require.NotNil(t, s2)
	require.NotEqual(t, s1, s2)

	require.Equal(t, s2, p.Get(ns2))
}
