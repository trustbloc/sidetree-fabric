/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package sidetreesvc

import (
	"testing"
	"time"

	cb "github.com/hyperledger/fabric-protos-go/common"
	"github.com/stretchr/testify/require"
	"github.com/trustbloc/fabric-peer-ext/pkg/gossip/blockpublisher"
	extmocks "github.com/trustbloc/fabric-peer-ext/pkg/mocks"
	extroles "github.com/trustbloc/fabric-peer-ext/pkg/roles"
	protocolApi "github.com/trustbloc/sidetree-core-go/pkg/api/protocol"
	"github.com/trustbloc/sidetree-core-go/pkg/batch/opqueue"
	coremocks "github.com/trustbloc/sidetree-core-go/pkg/mocks"

	"github.com/trustbloc/sidetree-fabric/pkg/config"
	cfgmocks "github.com/trustbloc/sidetree-fabric/pkg/config/mocks"
	sidetreectx "github.com/trustbloc/sidetree-fabric/pkg/context"
	"github.com/trustbloc/sidetree-fabric/pkg/mocks"
	"github.com/trustbloc/sidetree-fabric/pkg/observer"
	peermocks "github.com/trustbloc/sidetree-fabric/pkg/peer/mocks"
	"github.com/trustbloc/sidetree-fabric/pkg/rest/filehandler"
	"github.com/trustbloc/sidetree-fabric/pkg/rest/sidetreehandler"
	"github.com/trustbloc/sidetree-fabric/pkg/role"
)

//go:generate counterfeiter -o ../mocks/sidetreeconfigprovider.gen.go --fake-name SidetreeConfigProvider . sidetreeConfigProvider

const (
	channel1 = "channel1"
	channel2 = "channel2"
)

func TestProvider(t *testing.T) {
	rolesValue := make(map[extroles.Role]struct{})
	rolesValue[extroles.EndorserRole] = struct{}{}
	rolesValue[role.Resolver] = struct{}{}
	rolesValue[role.BatchWriter] = struct{}{}
	rolesValue[role.Observer] = struct{}{}
	extroles.SetRoles(rolesValue)
	defer func() {
		extroles.SetRoles(nil)
	}()

	peerConfig := &peermocks.PeerConfig{}
	peerConfig.MSPIDReturns(msp1)
	peerConfig.PeerIDReturns(peer1)

	sidetreeCfg := config.Sidetree{
		ChaincodeName:      "document",
		Collection:         "docs",
		BatchWriterTimeout: time.Second,
	}

	sidetreePeerCfg := config.SidetreePeer{}

	trustblocHandler := sidetreehandler.Config{
		Namespace: didTrustblocNamespace,
		BasePath:  didTrustblocBasePath,
	}

	pr := protocolApi.Protocol{
		GenesisTime:        100,
		MultihashAlgorithm: 18,
		MaxOperationCount:  100,
		MaxOperationSize:   1000,
	}

	protocolVersions := map[string]protocolApi.Protocol{"0.5": pr}

	pv := &coremocks.ProtocolVersion{}
	pv.ProtocolReturns(pr)

	vf := &peermocks.ProtocolVersionFactory{}
	vf.CreateProtocolVersionReturns(pv, nil)

	configSvc := &peermocks.ConfigService{}
	configProvider := &peermocks.ConfigServiceProvider{}
	configProvider.ForChannelReturns(configSvc)

	opQueueProvider := &mocks.OperationQueueProvider{}
	opQueueProvider.CreateReturns(&opqueue.MemQueue{}, nil)

	restConfig := &peermocks.RestConfig{}
	restConfig.SidetreeListenURLReturns("localhost:7721", nil)

	dcasClient := &mocks.DCASClient{}
	dcasProvider := &mocks.DCASClientProvider{}
	dcasProvider.GetDCASClientReturns(dcasClient, nil)

	observerProviders := &observer.ClientProviders{}
	gossip := extmocks.NewMockGossipAdapter()
	gossip.Self(msp1, extmocks.NewMember(peer1, []byte("pkiid")))

	gossipProvider := &extmocks.GossipProvider{}
	gossipProvider.GetGossipServiceReturns(gossip)

	observerProviders.Gossip = gossipProvider

	discoveryProvider := &peermocks.DiscoveryProvider{}

	ledgerProvider := &extmocks.LedgerProvider{}
	l := &extmocks.Ledger{
		BlockchainInfo: &cb.BlockchainInfo{
			Height: 1000,
		},
	}
	ledgerProvider.GetLedgerReturns(l)

	providers := &providers{
		ContextProviders: &ContextProviders{
			Providers: &sidetreectx.Providers{
				OperationQueueProvider: opQueueProvider,
				DCASProvider:           dcasProvider,
				LedgerProvider:         ledgerProvider,
			},
			VersionFactory: vf,
		},
		PeerConfig:        peerConfig,
		RESTConfig:        restConfig,
		ConfigProvider:    configProvider,
		BlockPublisher:    blockpublisher.NewProvider(),
		ObserverProviders: observerProviders,
		DiscoveryProvider: discoveryProvider,
	}

	fileHandler1 := filehandler.Config{
		BasePath:       "/path",
		ChaincodeName:  "cc1",
		Collection:     "coll1",
		IndexNamespace: "file:idx",
	}

	sidetreeCfgService2 := &cfgmocks.SidetreeConfigService{}
	sidetreeCfgService2.LoadSidetreeReturns(sidetreeCfg, nil)
	sidetreeCfgService2.LoadSidetreePeerReturns(sidetreePeerCfg, nil)
	sidetreeCfgService2.LoadSidetreeHandlersReturns([]sidetreehandler.Config{trustblocHandler}, nil)
	sidetreeCfgService2.LoadProtocolsReturns(protocolVersions, nil)
	sidetreeCfgService2.LoadFileHandlersReturns([]filehandler.Config{fileHandler1}, nil)

	sidetreeCfgService1 := &cfgmocks.SidetreeConfigService{}

	sidetreeCfgProvider := &peermocks.SidetreeConfigProvider{}
	sidetreeCfgProvider.ForChannelReturnsOnCall(0, sidetreeCfgService1)
	sidetreeCfgProvider.ForChannelReturnsOnCall(1, sidetreeCfgService2)
	sidetreeCfgProvider.ForChannelReturnsOnCall(2, sidetreeCfgService2)

	p := NewProvider(providers, sidetreeCfgProvider)
	require.NotNil(t, p)

	p.ChannelJoined(channel1)
	time.Sleep(20 * time.Millisecond)
	p.RestartRESTService()

	p.ChannelJoined(channel2)
	time.Sleep(20 * time.Millisecond)

	pcp, err := p.ProtocolClientProviderForChannel(channel1)
	require.NoError(t, err)
	require.NotNil(t, pcp)

	pcp, err = p.ProtocolClientProviderForChannel("invalid")
	require.EqualError(t, err, "protocol client provider not found for channel [invalid]")
	require.Nil(t, pcp)

	p.RestartRESTService()
	time.Sleep(20 * time.Millisecond)

	p.Close()
}
