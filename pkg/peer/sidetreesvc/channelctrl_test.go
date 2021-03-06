/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package sidetreesvc

import (
	"errors"
	"testing"
	"time"

	cb "github.com/hyperledger/fabric-protos-go/common"
	"github.com/stretchr/testify/require"
	olmocks "github.com/trustbloc/fabric-peer-ext/pkg/collections/offledger/mocks"
	ledgerconfig "github.com/trustbloc/fabric-peer-ext/pkg/config/ledgerconfig/config"
	cfgservice "github.com/trustbloc/fabric-peer-ext/pkg/config/ledgerconfig/service"
	extmocks "github.com/trustbloc/fabric-peer-ext/pkg/mocks"
	extroles "github.com/trustbloc/fabric-peer-ext/pkg/roles"
	protocolApi "github.com/trustbloc/sidetree-core-go/pkg/api/protocol"
	"github.com/trustbloc/sidetree-core-go/pkg/batch/opqueue"
	coremocks "github.com/trustbloc/sidetree-core-go/pkg/mocks"

	"github.com/trustbloc/sidetree-fabric/pkg/common"
	"github.com/trustbloc/sidetree-fabric/pkg/config"
	cfgmocks "github.com/trustbloc/sidetree-fabric/pkg/config/mocks"
	sidetreectx "github.com/trustbloc/sidetree-fabric/pkg/context"
	ctxmocks "github.com/trustbloc/sidetree-fabric/pkg/context/mocks"
	"github.com/trustbloc/sidetree-fabric/pkg/mocks"
	"github.com/trustbloc/sidetree-fabric/pkg/observer"
	obmocks "github.com/trustbloc/sidetree-fabric/pkg/observer/mocks"
	peerconfig "github.com/trustbloc/sidetree-fabric/pkg/peer/config"
	"github.com/trustbloc/sidetree-fabric/pkg/peer/discovery"
	peermocks "github.com/trustbloc/sidetree-fabric/pkg/peer/mocks"
	"github.com/trustbloc/sidetree-fabric/pkg/rest/authhandler"
	"github.com/trustbloc/sidetree-fabric/pkg/rest/blockchainhandler"
	"github.com/trustbloc/sidetree-fabric/pkg/rest/dcashandler"
	"github.com/trustbloc/sidetree-fabric/pkg/rest/discoveryhandler"
	"github.com/trustbloc/sidetree-fabric/pkg/rest/filehandler"
	"github.com/trustbloc/sidetree-fabric/pkg/rest/sidetreehandler"
	"github.com/trustbloc/sidetree-fabric/pkg/role"
)

//go:generate counterfeiter -o ../mocks/restservercontroller.gen.go --fake-name RESTServerController . restServiceController
//go:generate counterfeiter -o ../mocks/configserviceprovider.gen.go --fake-name ConfigServiceProvider . configServiceProvider
//go:generate counterfeiter -o ../mocks/discoveryprovider.gen.go --fake-name DiscoveryProvider . discoveryProvider
//go:generate counterfeiter -o ../mocks/configservice.gen.go --fake-name ConfigService github.com/trustbloc/fabric-peer-ext/pkg/config/ledgerconfig/config.Service

const (
	peer1Address = "peer1.example.com:80"
	domain1      = "example.com"

	didTrustblocNamespace = "did:bloc:trustbloc.dev"
	didTrustblocBasePath  = "/trustbloc.dev"

	fileIndexNamespace = "file:idx"
	fileIndexBasePath  = "/file"

	eventMethod = "RestartRESTService"

	tx1 = "tx1"
	tx2 = "tx2"
)

func TestChannelController_Update(t *testing.T) {
	restoreRoles := setRoles(role.Resolver, role.BatchWriter, role.Observer)
	defer restoreRoles()

	sidetreePeerCfg := config.SidetreePeer{}

	sidetreeHandlers := []sidetreehandler.Config{
		{
			Namespace: didTrustblocNamespace,
			BasePath:  didTrustblocBasePath,
			Authorization: authhandler.Config{
				ReadTokens:  []string{"did_r", "did_w"},
				WriteTokens: []string{"did_w"},
			},
		},
		{
			Namespace: fileIndexNamespace,
			BasePath:  fileIndexBasePath,
			DocType:   common.FileIndexType,
			Authorization: authhandler.Config{
				ReadTokens:  []string{"content_r"},
				WriteTokens: []string{"content_r", "content_w"},
			},
		},
	}

	p := protocolApi.Protocol{
		GenesisTime:         100,
		MultihashAlgorithms: []uint{18},
		MaxOperationCount:   100,
		MaxOperationSize:    1000,
	}

	protocolVersions := map[string]protocolApi.Protocol{"0.5": p}

	fileHandlers := []filehandler.Config{
		{
			BasePath:       "/schema",
			ChaincodeName:  "files",
			Collection:     "schemas",
			IndexNamespace: "file:idx",
			IndexDocID:     "file:idx:1234",
			Authorization: authhandler.Config{
				ReadTokens:  []string{"content_r"},
				WriteTokens: []string{"content_r", "content_w"},
			},
		},
	}

	dcasHandlers := []dcashandler.Config{
		{
			BasePath:      "/cas",
			ChaincodeName: "cascc",
			Collection:    "cas",
			Authorization: authhandler.Config{
				ReadTokens:  []string{"cas_r", "cas_w"},
				WriteTokens: []string{"cas_w"},
			},
		},
	}

	peerConfig := &peermocks.PeerConfig{}
	peerConfig.MSPIDReturns(msp1)
	peerConfig.PeerIDReturns(peer1)
	peerConfig.PeerAddressReturns(peer1Address)

	restCfg := &peermocks.RestConfig{}

	configSvc := &peermocks.ConfigService{}
	configProvider := &peermocks.ConfigServiceProvider{}
	configProvider.ForChannelReturns(configSvc)

	opQueue := &opqueue.MemQueue{}
	opQueueProvider := &mocks.OperationQueueProvider{}
	opQueueProvider.CreateReturns(opQueue, nil)

	dcas := &mocks.DCASClient{}
	dcasProvider := &mocks.DCASClientProvider{}
	dcasProvider.GetDCASClientReturns(dcas, nil)

	olClient := &olmocks.OffLedgerClient{}
	olProvider := &obmocks.OffLedgerClientProvider{}
	olProvider.ForChannelReturns(olClient, nil)

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

	v := &coremocks.ProtocolVersion{}
	v.ProtocolReturns(p)

	vf := &peermocks.ProtocolVersionFactory{}
	vf.CreateProtocolVersionReturns(v, nil)

	cacheProvider := &ctxmocks.CachingOpProcessorProvider{}

	providers := &providers{
		ContextProviders: &ContextProviders{
			Providers: &sidetreectx.Providers{
				DCASProvider:               dcasProvider,
				OffLedgerProvider:          olProvider,
				OperationQueueProvider:     opQueueProvider,
				LedgerProvider:             ledgerProvider,
				OperationProcessorProvider: cacheProvider,
			},
			VersionFactory: vf,
		},
		PeerConfig:        peerConfig,
		ConfigProvider:    configProvider,
		BlockPublisher:    extmocks.NewBlockPublisherProvider(),
		RESTConfig:        restCfg,
		ObserverProviders: observerProviders,
		DiscoveryProvider: discoveryProvider,
	}

	stConfigService := &cfgmocks.SidetreeConfigService{}
	stConfigService.LoadProtocolsReturns(protocolVersions, nil)
	stConfigService.LoadFileHandlersReturns(fileHandlers, nil)
	stConfigService.LoadDCASHandlersReturns(dcasHandlers, nil)
	stConfigService.LoadSidetreeHandlersReturns(sidetreeHandlers, nil)

	ctrl := &peermocks.RESTServerController{}

	m := newChannelController(channel1, providers, stConfigService, ctrl)
	require.NotNil(t, m)

	// config has not been loaded yet to so error is expected
	pc, err := m.ForNamespace(didTrustblocNamespace)
	require.Error(t, err)
	require.Contains(t, err.Error(), "protocol: context not found for namespace")

	defer m.Close()

	count := len(ctrl.Invocations()[eventMethod])
	stConfigService.LoadSidetreePeerReturns(sidetreePeerCfg, nil)
	require.NoError(t, m.load())

	// config has been loaded for namespace - protocol client is available
	pc, err = m.ForNamespace(didTrustblocNamespace)
	require.NoError(t, err)
	require.NotNil(t, pc)

	time.Sleep(20 * time.Millisecond)
	require.Len(t, ctrl.Invocations()[eventMethod], count+1)
	require.Len(t, m.RESTHandlers(), 11)

	localServices := m.localServices()
	require.Len(t, localServices, 4)

	serviceMap := make(map[string]discovery.Service)

	for _, s := range localServices {
		serviceMap[s.Service] = s
	}

	require.Len(t, serviceMap, 4)
	s, ok := serviceMap["cas"]
	require.True(t, ok)
	require.Equal(t, domain1, s.Domain)
	require.Equal(t, apiVersion, s.APIVersion)

	s, ok = serviceMap["schema"]
	require.True(t, ok)
	require.Equal(t, domain1, s.Domain)
	require.Equal(t, "", s.APIVersion)

	s, ok = serviceMap[didTrustblocNamespace]
	require.True(t, ok)
	require.Equal(t, domain1, s.Domain)
	require.Equal(t, apiVersion, s.APIVersion)

	s, ok = serviceMap[fileIndexNamespace]
	require.True(t, ok)
	require.Equal(t, domain1, s.Domain)
	require.Equal(t, apiVersion, s.APIVersion)

	t.Run("Update peer config -> success", func(t *testing.T) {
		count := len(ctrl.Invocations()[eventMethod])
		m.handleUpdate(&ledgerconfig.KeyValue{
			Key:   ledgerconfig.NewPeerKey(msp1, peer1, peerconfig.SidetreePeerAppName, peerconfig.SidetreePeerAppVersion),
			Value: &ledgerconfig.Value{TxID: tx1},
		})

		time.Sleep(100 * time.Millisecond)
		require.Len(t, ctrl.Invocations()[eventMethod], count+1)
	})

	t.Run("Update consortium config -> success", func(t *testing.T) {
		count := len(ctrl.Invocations()[eventMethod])
		m.handleUpdate(&ledgerconfig.KeyValue{
			Key:   ledgerconfig.NewAppKey(peerconfig.GlobalMSPID, didTrustblocNamespace, "1"),
			Value: &ledgerconfig.Value{TxID: tx2},
		})

		time.Sleep(100 * time.Millisecond)
		require.Len(t, ctrl.Invocations()[eventMethod], count+1)

		count = len(ctrl.Invocations()[eventMethod])
		m.handleUpdate(&ledgerconfig.KeyValue{
			Key:   ledgerconfig.NewAppKey(peerconfig.GlobalMSPID, didTrustblocNamespace, "1"),
			Value: &ledgerconfig.Value{TxID: tx2},
		})

		time.Sleep(100 * time.Millisecond)
		require.Len(t, ctrl.Invocations()[eventMethod], count)
	})

	t.Run("Irrelevant config update -> success", func(t *testing.T) {
		count := len(ctrl.Invocations()[eventMethod])

		m.handleUpdate(&ledgerconfig.KeyValue{
			Key:   ledgerconfig.NewAppKey(peerconfig.GlobalMSPID, "some-app-name", "1"),
			Value: &ledgerconfig.Value{TxID: tx1},
		})

		time.Sleep(100 * time.Millisecond)
		require.Len(t, ctrl.Invocations()[eventMethod], count)
	})

	t.Run("Peer config not found", func(t *testing.T) {
		stConfigService := &cfgmocks.SidetreeConfigService{}
		stConfigService.LoadSidetreePeerReturns(config.SidetreePeer{}, cfgservice.ErrConfigNotFound)

		m := newChannelController(channel1, providers, stConfigService, ctrl)
		require.NotNil(t, m)
		defer m.Close()

		count := len(ctrl.Invocations()[eventMethod])

		m.handleUpdate(&ledgerconfig.KeyValue{
			Key:   ledgerconfig.NewAppKey(peerconfig.GlobalMSPID, didTrustblocNamespace, "1"),
			Value: &ledgerconfig.Value{TxID: tx1},
		})

		time.Sleep(100 * time.Millisecond)
		require.Len(t, ctrl.Invocations()[eventMethod], count)
	})

	t.Run("File handler config not found", func(t *testing.T) {
		stConfigService := &cfgmocks.SidetreeConfigService{}
		stConfigService.LoadFileHandlersReturns(nil, cfgservice.ErrConfigNotFound)

		m := newChannelController(channel1, providers, stConfigService, ctrl)
		require.NotNil(t, m)
		defer m.Close()

		count := len(ctrl.Invocations()[eventMethod])

		m.handleUpdate(&ledgerconfig.KeyValue{
			Key:   ledgerconfig.NewPeerKey(msp1, peer1, peerconfig.FileHandlerAppName, peerconfig.FileHandlerAppVersion),
			Value: &ledgerconfig.Value{TxID: tx1},
		})

		time.Sleep(100 * time.Millisecond)
		require.Len(t, ctrl.Invocations()[eventMethod], count+1)
	})

	t.Run("Sidetree handler config not found", func(t *testing.T) {
		stConfigService := &cfgmocks.SidetreeConfigService{}
		stConfigService.LoadSidetreeHandlersReturns(nil, cfgservice.ErrConfigNotFound)

		m := newChannelController(channel1, providers, stConfigService, ctrl)
		require.NotNil(t, m)
		defer m.Close()

		count := len(ctrl.Invocations()[eventMethod])

		m.handleUpdate(&ledgerconfig.KeyValue{
			Key:   ledgerconfig.NewPeerKey(msp1, peer1, peerconfig.FileHandlerAppName, peerconfig.FileHandlerAppVersion),
			Value: &ledgerconfig.Value{TxID: tx1},
		})

		time.Sleep(100 * time.Millisecond)
		require.Len(t, ctrl.Invocations()[eventMethod], count+1)
	})
}

func TestChannelController_LoadDCASHandlers(t *testing.T) {
	peerConfig := &peermocks.PeerConfig{}
	peerConfig.MSPIDReturns(msp1)
	peerConfig.PeerIDReturns(peer1)

	restCfg := &peermocks.RestConfig{}

	configSvc := &peermocks.ConfigService{}
	configProvider := &peermocks.ConfigServiceProvider{}
	configProvider.ForChannelReturns(configSvc)

	opQueue := &opqueue.MemQueue{}
	opQueueProvider := &mocks.OperationQueueProvider{}
	opQueueProvider.CreateReturns(opQueue, nil)

	dcas := &mocks.DCASClient{}
	dcasProvider := &mocks.DCASClientProvider{}
	dcasProvider.GetDCASClientReturns(dcas, nil)

	olClient := &olmocks.OffLedgerClient{}
	olProvider := &obmocks.OffLedgerClientProvider{}
	olProvider.ForChannelReturns(olClient, nil)

	discoveryProvider := &peermocks.DiscoveryProvider{}

	ledgerProvider := &extmocks.LedgerProvider{}
	l := &extmocks.Ledger{
		BlockchainInfo: &cb.BlockchainInfo{
			Height: 1000,
		},
	}
	ledgerProvider.GetLedgerReturns(l)

	cacheProvider := &ctxmocks.CachingOpProcessorProvider{}

	providers := &providers{
		ContextProviders: &ContextProviders{
			Providers: &sidetreectx.Providers{
				DCASProvider:               dcasProvider,
				OffLedgerProvider:          olProvider,
				OperationQueueProvider:     opQueueProvider,
				LedgerProvider:             ledgerProvider,
				OperationProcessorProvider: cacheProvider,
			},
		},
		PeerConfig:        peerConfig,
		ConfigProvider:    configProvider,
		BlockPublisher:    extmocks.NewBlockPublisherProvider(),
		RESTConfig:        restCfg,
		DiscoveryProvider: discoveryProvider,
	}

	stConfigService := &cfgmocks.SidetreeConfigService{}

	ctrl := &peermocks.RESTServerController{}

	c := newChannelController(channel1, providers, stConfigService, ctrl)
	require.NotNil(t, c)

	defer c.Close()

	// No config
	stConfigService.LoadSidetreePeerReturns(config.SidetreePeer{}, cfgservice.ErrConfigNotFound)
	require.NoError(t, c.load())
	require.Empty(t, c.RESTHandlers())

	// Just DCAS handlers
	dcasHandlers := []dcashandler.Config{
		{
			BasePath:      "/cas",
			ChaincodeName: "cascc",
			Collection:    "cas",
		},
	}

	stConfigService.LoadDCASHandlersReturns(dcasHandlers, nil)
	require.NoError(t, c.load())
	require.Len(t, c.RESTHandlers(), 3)
}

func TestChannelController_LoadBlockchainHandlers(t *testing.T) {
	peerConfig := &peermocks.PeerConfig{}
	peerConfig.MSPIDReturns(msp1)
	peerConfig.PeerIDReturns(peer1)

	restCfg := &peermocks.RestConfig{}

	configSvc := &peermocks.ConfigService{}
	configProvider := &peermocks.ConfigServiceProvider{}
	configProvider.ForChannelReturns(configSvc)

	opQueue := &opqueue.MemQueue{}
	opQueueProvider := &mocks.OperationQueueProvider{}
	opQueueProvider.CreateReturns(opQueue, nil)

	blockchainProvider := &obmocks.BlockchainClientProvider{}

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
				LedgerProvider:         ledgerProvider,
			},
			BlockchainProvider: blockchainProvider,
		},
		PeerConfig:        peerConfig,
		ConfigProvider:    configProvider,
		BlockPublisher:    extmocks.NewBlockPublisherProvider(),
		RESTConfig:        restCfg,
		DiscoveryProvider: discoveryProvider,
	}

	stConfigService := &cfgmocks.SidetreeConfigService{}

	ctrl := &peermocks.RESTServerController{}

	c := newChannelController(channel1, providers, stConfigService, ctrl)
	require.NotNil(t, c)

	defer c.Close()

	// No config
	stConfigService.LoadSidetreePeerReturns(config.SidetreePeer{}, cfgservice.ErrConfigNotFound)
	require.NoError(t, c.load())
	require.Empty(t, c.RESTHandlers())

	// Just blockchain handlers
	blockchainHandlers := []blockchainhandler.Config{
		{
			BasePath: "/blockchain",
			Authorization: authhandler.Config{
				ReadTokens:  []string{"blockchain_r", "blockchain_w"},
				WriteTokens: []string{"blockchain_w"},
			},
		},
	}

	stConfigService.LoadBlockchainHandlersReturns(nil, errors.New("injected error"))
	require.Error(t, c.load())

	stConfigService.LoadBlockchainHandlersReturns(nil, cfgservice.ErrConfigNotFound)
	require.NoError(t, c.load())

	stConfigService.LoadBlockchainHandlersReturns(blockchainHandlers, nil)
	require.NoError(t, c.load())
	require.Len(t, c.RESTHandlers(), 14)
}

func TestChannelController_LoadDiscoveryHandlers(t *testing.T) {
	peerConfig := &peermocks.PeerConfig{}
	peerConfig.MSPIDReturns(msp1)
	peerConfig.PeerIDReturns(peer1)

	restCfg := &peermocks.RestConfig{}

	configSvc := &peermocks.ConfigService{}
	configProvider := &peermocks.ConfigServiceProvider{}
	configProvider.ForChannelReturns(configSvc)

	opQueue := &opqueue.MemQueue{}
	opQueueProvider := &mocks.OperationQueueProvider{}
	opQueueProvider.CreateReturns(opQueue, nil)

	blockchainProvider := &obmocks.BlockchainClientProvider{}

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
				LedgerProvider:         ledgerProvider,
			},
			BlockchainProvider: blockchainProvider,
		},
		PeerConfig:        peerConfig,
		ConfigProvider:    configProvider,
		BlockPublisher:    extmocks.NewBlockPublisherProvider(),
		RESTConfig:        restCfg,
		DiscoveryProvider: discoveryProvider,
	}

	stConfigService := &cfgmocks.SidetreeConfigService{}

	ctrl := &peermocks.RESTServerController{}

	c := newChannelController(channel1, providers, stConfigService, ctrl)
	require.NotNil(t, c)

	defer c.Close()

	// No config
	stConfigService.LoadSidetreePeerReturns(config.SidetreePeer{}, cfgservice.ErrConfigNotFound)
	require.NoError(t, c.load())
	require.Empty(t, c.RESTHandlers())

	// Just discovery handlers
	discoveryHandlers := []discoveryhandler.Config{
		{
			BasePath: "/discovery",
			Authorization: authhandler.Config{
				ReadTokens: []string{"discovery_r"},
			},
		},
	}

	stConfigService.LoadDiscoveryHandlersReturns(nil, errors.New("injected error"))
	require.Error(t, c.load())

	stConfigService.LoadDiscoveryHandlersReturns(nil, cfgservice.ErrConfigNotFound)
	require.NoError(t, c.load())

	stConfigService.LoadDiscoveryHandlersReturns(discoveryHandlers, nil)
	require.NoError(t, c.load())
	require.Len(t, c.RESTHandlers(), 1)
}

func setRoles(roles ...extroles.Role) func() {
	rolesValue := make(map[extroles.Role]struct{})

	for _, r := range roles {
		rolesValue[r] = struct{}{}
	}

	extroles.SetRoles(rolesValue)

	return func() {
		extroles.SetRoles(nil)
	}
}
