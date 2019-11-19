/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package peer

import (
	"testing"

	"github.com/stretchr/testify/require"
	clientmocks "github.com/trustbloc/fabric-peer-ext/pkg/collections/client/mocks"
	"github.com/trustbloc/fabric-peer-ext/pkg/mocks"
	extpeer "github.com/trustbloc/fabric-peer-ext/pkg/peer"
	"github.com/trustbloc/fabric-peer-ext/pkg/resource"
)

func TestInitialize(t *testing.T) {
	require.NotPanics(t, extpeer.Initialize)
	require.NotPanics(t, Initialize)

	require.NoError(t, resource.Mgr.Initialize(
		mocks.NewBlockPublisherProvider(),
		&mocks.LedgerProvider{},
		mocks.NewMockGossipAdapter(),
		&clientmocks.PvtDataDistributor{},
		&mocks.IdentityDeserializerProvider{},
		&mocks.IdentifierProvider{},
		&mocks.IdentityProvider{},
	))

	require.NotPanics(t, resource.Mgr.Close)
}
