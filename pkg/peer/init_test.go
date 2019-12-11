/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package peer

import (
	"os"
	"testing"
	"time"

	cb "github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric/extensions/collections/storeprovider"
	"github.com/hyperledger/fabric/extensions/gossip/blockpublisher"
	"github.com/stretchr/testify/require"
	"github.com/trustbloc/fabric-peer-ext/pkg/config"
	statemocks "github.com/trustbloc/fabric-peer-ext/pkg/gossip/state/mocks"
	"github.com/trustbloc/fabric-peer-ext/pkg/mocks"
	extpeer "github.com/trustbloc/fabric-peer-ext/pkg/peer"
	"github.com/trustbloc/fabric-peer-ext/pkg/resource"
	observercfg "github.com/trustbloc/sidetree-fabric/pkg/observer/config"
)

const (
	channelID = "testchannel"
	peerID    = "peer1.example.com"
)

func TestInitialize(t *testing.T) {
	defer removeDBPath(t)

	// Ensure that the provider instances are instantiated and registered as a resource
	require.NotNil(t, blockpublisher.ProviderInstance)
	require.NotNil(t, storeprovider.NewProviderFactory())

	require.NotPanics(t, extpeer.Initialize)
	require.NotPanics(t, Initialize)

	lp := &mocks.LedgerProvider{}
	l := &mocks.Ledger{
		BlockchainInfo: &cb.BlockchainInfo{
			Height: 1000,
		},
	}
	lp.GetLedgerReturns(l)

	require.NoError(t, resource.Mgr.Initialize(
		mocks.NewBlockPublisherProvider(),
		lp,
		&mocks.GossipProvider{},
		&mocks.IdentityDeserializerProvider{},
		&mocks.IdentifierProvider{},
		&mocks.IdentityProvider{},
		&statemocks.CCEventMgrProvider{},
		observercfg.New(peerID, []string{channelID}, time.Second),
	))

	require.NotPanics(t, func() { resource.Mgr.ChannelJoined(channelID) })
	require.NotPanics(t, resource.Mgr.Close)
}

func removeDBPath(t testing.TB) {
	removePath(t, config.GetTransientDataLevelDBPath())
}

func removePath(t testing.TB, path string) {
	if err := os.RemoveAll(path); err != nil {
		t.Fatalf(err.Error())
	}
}
