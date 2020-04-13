/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package sidetreesvc

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	ledgerconfig "github.com/trustbloc/fabric-peer-ext/pkg/config/ledgerconfig/config"
	"github.com/trustbloc/fabric-peer-ext/pkg/config/ledgerconfig/service"
	extmocks "github.com/trustbloc/fabric-peer-ext/pkg/mocks"
	extroles "github.com/trustbloc/fabric-peer-ext/pkg/roles"
	protocolApi "github.com/trustbloc/sidetree-core-go/pkg/api/protocol"
	"github.com/trustbloc/sidetree-core-go/pkg/batch/opqueue"

	"github.com/trustbloc/sidetree-fabric/pkg/config"
	cfgmocks "github.com/trustbloc/sidetree-fabric/pkg/config/mocks"
	"github.com/trustbloc/sidetree-fabric/pkg/filehandler"
	"github.com/trustbloc/sidetree-fabric/pkg/mocks"
	peerconfig "github.com/trustbloc/sidetree-fabric/pkg/peer/config"
	peermocks "github.com/trustbloc/sidetree-fabric/pkg/peer/mocks"
	"github.com/trustbloc/sidetree-fabric/pkg/rest/dcashandler"
	"github.com/trustbloc/sidetree-fabric/pkg/role"
)

//go:generate counterfeiter -o ../mocks/restservercontroller.gen.go --fake-name RESTServerController . restServiceController
//go:generate counterfeiter -o ../mocks/configserviceprovider.gen.go --fake-name ConfigServiceProvider . configServiceProvider
//go:generate counterfeiter -o ../mocks/configservice.gen.go --fake-name ConfigService github.com/trustbloc/fabric-peer-ext/pkg/config/ledgerconfig/config.Service

const (
	didTrustblocNamespace = "did:bloc:trustbloc.dev"
	didTrustblocBasePath  = "/trustbloc.dev"

	fileIndexNamespace = "file:idx"
	fileIndexBasePath  = "/file"

	eventMethod = "RestartRESTService"

	tx1 = "tx1"
	tx2 = "tx2"
)

func TestChannelManager(t *testing.T) {
	rolesValue := make(map[extroles.Role]struct{})
	rolesValue[extroles.EndorserRole] = struct{}{}
	rolesValue[role.Resolver] = struct{}{}
	rolesValue[role.BatchWriter] = struct{}{}
	rolesValue[role.Observer] = struct{}{}
	extroles.SetRoles(rolesValue)
	defer func() {
		extroles.SetRoles(nil)
	}()

	sidetreePeerCfg := config.SidetreePeer{}
	sidetreePeerCfg.Namespaces = []config.Namespace{
		{
			Namespace: didTrustblocNamespace,
			BasePath:  didTrustblocBasePath,
		},
		{
			Namespace: fileIndexNamespace,
			BasePath:  fileIndexBasePath,
			DocType:   config.FileIndexType,
		},
	}

	protocolVersions := map[string]protocolApi.Protocol{
		"0.5": {
			StartingBlockChainTime:       100,
			HashAlgorithmInMultiHashCode: 18,
			MaxOperationsPerBatch:        100,
			MaxOperationByteSize:         1000,
		},
	}

	fileHandlers := []filehandler.Config{
		{
			BasePath:       "/schema",
			ChaincodeName:  "files",
			Collection:     "schemas",
			IndexNamespace: "file:idx",
			IndexDocID:     "file:idx:1234",
		},
	}

	dcasHandlers := []dcashandler.Config{
		{
			BasePath:      "/cas",
			ChaincodeName: "cascc",
			Collection:    "cas",
		},
	}

	peerConfig := &peermocks.PeerConfig{}
	peerConfig.MSPIDReturns(msp1)
	peerConfig.PeerIDReturns(peer1)

	configSvc := &peermocks.ConfigService{}
	configProvider := &peermocks.ConfigServiceProvider{}
	configProvider.ForChannelReturns(configSvc)

	opQueue := &opqueue.MemQueue{}
	opQueueProvider := &mocks.OperationQueueProvider{}
	opQueueProvider.CreateReturns(opQueue, nil)

	dcas := &mocks.DCASClient{}
	dcasProvider := &mocks.DCASClientProvider{}
	dcasProvider.ForChannelReturns(dcas, nil)

	providers := &providers{
		ContextProviders: &ContextProviders{
			DCASProvider:           dcasProvider,
			OperationQueueProvider: opQueueProvider,
		},
		PeerConfig:     peerConfig,
		ConfigProvider: configProvider,
		BlockPublisher: extmocks.NewBlockPublisherProvider(),
	}

	stConfigService := &cfgmocks.SidetreeConfigService{}
	stConfigService.LoadProtocolsReturns(protocolVersions, nil)
	stConfigService.LoadFileHandlersReturns(fileHandlers, nil)
	stConfigService.LoadDCASHandlersReturns(dcasHandlers, nil)

	ctrl := &peermocks.RESTServerController{}

	m := newChannelController(channel1, providers, stConfigService, ctrl)
	require.NotNil(t, m)

	defer m.Close()

	count := len(ctrl.Invocations()[eventMethod])
	stConfigService.LoadSidetreePeerReturns(sidetreePeerCfg, nil)
	require.NoError(t, m.load())

	time.Sleep(20 * time.Millisecond)
	require.Len(t, ctrl.Invocations()[eventMethod], count+1)
	require.Len(t, m.RESTHandlers(), 6)

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
		stConfigService.LoadSidetreePeerReturns(config.SidetreePeer{}, service.ErrConfigNotFound)

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
		stConfigService.LoadFileHandlersReturns(nil, service.ErrConfigNotFound)

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
