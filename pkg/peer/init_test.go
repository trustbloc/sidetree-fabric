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
	viper "github.com/spf13/viper2015"
	"github.com/stretchr/testify/require"
	"github.com/trustbloc/fabric-peer-ext/pkg/config"
	statemocks "github.com/trustbloc/fabric-peer-ext/pkg/gossip/state/mocks"
	"github.com/trustbloc/fabric-peer-ext/pkg/mocks"
	extpeer "github.com/trustbloc/fabric-peer-ext/pkg/peer"
	"github.com/trustbloc/fabric-peer-ext/pkg/resource"
	extroles "github.com/trustbloc/fabric-peer-ext/pkg/roles"
	observercfg "github.com/trustbloc/sidetree-fabric/pkg/observer/config"
	"github.com/trustbloc/sidetree-fabric/pkg/role"
)

const (
	channelID          = "mychannel"
	peerID             = "peer1.example.com"
	configFile         = "../context/testdata/config.yaml"
	protocolConfigFile = "../context/testdata/protocol.json"
	keyProtocolFile    = "sidetree.protocol.file"
)

func TestInitialize(t *testing.T) {
	defer removeDBPath(t)

	viper.Set(keyConfigFile, configFile)
	viper.Set(keyProtocolFile, protocolConfigFile)

	// Ensure that the provider instances are instantiated and registered as a resource
	require.NotNil(t, blockpublisher.ProviderInstance)
	require.NotNil(t, storeprovider.NewProviderFactory())
	require.NotNil(t, extroles.GetRoles())

	rolesValue := make(map[extroles.Role]struct{})
	rolesValue[role.Observer] = struct{}{}
	rolesValue[role.BatchWriter] = struct{}{}
	rolesValue[role.Resolver] = struct{}{}
	extroles.SetRoles(rolesValue)
	defer func() {
		extroles.SetRoles(nil)
	}()

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
		&mockBatchWriterConfig{batchTimeout: time.Second},
		&mockRESTServiceConfig{listenURL: "localhost:8978"},
	))

	require.NotPanics(t, func() { resource.Mgr.ChannelJoined(channelID) })

	// Give the services a chance to startup
	time.Sleep(2 * time.Second)

	require.NotPanics(t, resource.Mgr.Close)
}

func TestSidetreeConfig_ChannelID(t *testing.T) {
	viper.Set(keyConfigFile, configFile)

	c := newSidetreeConfigResource()
	require.NotNil(t, c)
	require.Equal(t, "mychannel", c.ChannelID())
}

func removeDBPath(t testing.TB) {
	removePath(t, config.GetTransientDataLevelDBPath())
}

func removePath(t testing.TB, path string) {
	if err := os.RemoveAll(path); err != nil {
		t.Fatalf(err.Error())
	}
}

type mockBatchWriterConfig struct {
	batchTimeout time.Duration
}

func (m *mockBatchWriterConfig) GetBatchTimeout() time.Duration {
	return m.batchTimeout
}

type mockRESTServiceConfig struct {
	listenURL string
}

func (m *mockRESTServiceConfig) GetListenURL() string {
	return m.listenURL
}
