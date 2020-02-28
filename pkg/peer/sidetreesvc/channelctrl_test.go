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

	"github.com/trustbloc/sidetree-fabric/pkg/mocks"
	"github.com/trustbloc/sidetree-fabric/pkg/observer"
	"github.com/trustbloc/sidetree-fabric/pkg/peer/config"
	peermocks "github.com/trustbloc/sidetree-fabric/pkg/peer/mocks"
	"github.com/trustbloc/sidetree-fabric/pkg/role"
)

//go:generate counterfeiter -o ../mocks/restservercontroller.gen.go --fake-name RESTServerController . restServiceController
//go:generate counterfeiter -o ../mocks/configserviceprovider.gen.go --fake-name ConfigServiceProvider . configServiceProvider
//go:generate counterfeiter -o ../mocks/configservice.gen.go --fake-name ConfigService github.com/trustbloc/fabric-peer-ext/pkg/config/ledgerconfig/config.Service

const (
	didTrustblocNamespace = "did:bloc:trustbloc.dev"
	didTrustblocBasePath  = "/trustbloc.dev"

	eventMethod = "RestartRESTService"
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
	}

	protocolVersions := map[string]protocolApi.Protocol{
		"0.5": {
			StartingBlockChainTime:       100,
			HashAlgorithmInMultiHashCode: 18,
			MaxOperationsPerBatch:        100,
			MaxOperationByteSize:         1000,
		},
	}

	peerConfig := &peermocks.PeerConfig{}
	peerConfig.MSPIDReturns(msp1)
	peerConfig.PeerIDReturns(peer1)

	configSvc := &peermocks.ConfigService{}
	configProvider := &peermocks.ConfigServiceProvider{}
	configProvider.ForChannelReturns(configSvc)

	observerProviders := &observer.Providers{
		BlockPublisher: extmocks.NewBlockPublisherProvider(),
	}

	opQueue := &opqueue.MemQueue{}
	opQueueProvider := &mocks.OperationQueueProvider{}
	opQueueProvider.CreateReturns(opQueue, nil)

	providers := &providers{
		PeerConfig:             peerConfig,
		ConfigProvider:         configProvider,
		ObserverProviders:      observerProviders,
		OperationQueueProvider: opQueueProvider,
	}

	stConfigService := &peermocks.SidetreeConfigService{}
	stConfigService.LoadProtocolsReturns(protocolVersions, nil)

	ctrl := &peermocks.RESTServerController{}

	m := newChannelController(channel1, providers, stConfigService, ctrl)
	require.NotNil(t, m)

	defer m.Close()

	count := len(ctrl.Invocations()[eventMethod])
	stConfigService.LoadSidetreePeerReturns(sidetreePeerCfg, nil)
	require.NoError(t, m.load())

	time.Sleep(20 * time.Millisecond)
	require.Len(t, ctrl.Invocations()[eventMethod], count+1)
	require.Len(t, m.RESTHandlers(), 2)

	t.Run("Update peer sidetreeCfgService -> success", func(t *testing.T) {
		count := len(ctrl.Invocations()[eventMethod])
		m.handleUpdate(&ledgerconfig.KeyValue{
			Key: ledgerconfig.NewPeerKey(msp1, peer1, config.SidetreePeerAppName, config.SidetreePeerAppVersion),
		})

		time.Sleep(20 * time.Millisecond)
		require.Len(t, ctrl.Invocations()[eventMethod], count+1)
	})

	t.Run("Update consortium sidetreeCfgService -> success", func(t *testing.T) {
		count := len(ctrl.Invocations()[eventMethod])
		m.handleUpdate(&ledgerconfig.KeyValue{
			Key: ledgerconfig.NewAppKey(config.GlobalMSPID, didTrustblocNamespace, "1"),
		})

		time.Sleep(20 * time.Millisecond)
		require.Len(t, ctrl.Invocations()[eventMethod], count+1)
	})

	t.Run("Irrelevant sidetreeCfgService update -> success", func(t *testing.T) {
		count := len(ctrl.Invocations()[eventMethod])

		m.handleUpdate(&ledgerconfig.KeyValue{
			Key: ledgerconfig.NewAppKey(config.GlobalMSPID, "some-app-name", "1"),
		})

		time.Sleep(20 * time.Millisecond)
		require.Len(t, ctrl.Invocations()[eventMethod], count)
	})

	t.Run("Peer sidetreeCfgService not found", func(t *testing.T) {
		count := len(ctrl.Invocations()[eventMethod])

		stConfigService.LoadSidetreePeerReturns(config.SidetreePeer{}, service.ErrConfigNotFound)
		m.handleUpdate(&ledgerconfig.KeyValue{
			Key: ledgerconfig.NewAppKey(config.GlobalMSPID, didTrustblocNamespace, "1"),
		})

		time.Sleep(20 * time.Millisecond)
		require.Len(t, ctrl.Invocations()[eventMethod], count)
	})
}
