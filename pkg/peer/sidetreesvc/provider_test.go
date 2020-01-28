/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package sidetreesvc

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	extmocks "github.com/trustbloc/fabric-peer-ext/pkg/mocks"
	extroles "github.com/trustbloc/fabric-peer-ext/pkg/roles"
	protocolApi "github.com/trustbloc/sidetree-core-go/pkg/api/protocol"
	"github.com/trustbloc/sidetree-fabric/pkg/observer"
	"github.com/trustbloc/sidetree-fabric/pkg/peer/config"
	"github.com/trustbloc/sidetree-fabric/pkg/peer/mocks"
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

	peerConfig := &mocks.PeerConfig{}
	peerConfig.MSPIDReturns(msp1)
	peerConfig.PeerIDReturns(peer1)

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

	configSvc := &mocks.ConfigService{}
	configProvider := &mocks.ConfigServiceProvider{}
	configProvider.ForChannelReturns(configSvc)

	observerProviders := &observer.Providers{
		BlockPublisher: extmocks.NewBlockPublisherProvider(),
	}

	restConfig := &mocks.RestConfig{}
	restConfig.SidetreeListenURLReturns("localhost:7721", nil)

	providers := &providers{
		PeerConfig:        peerConfig,
		ConfigProvider:    configProvider,
		ObserverProviders: observerProviders,
		RESTConfig:        restConfig,
	}

	sidetreeCfgService2 := &mocks.SidetreeConfigService{}
	sidetreeCfgService2.LoadSidetreePeerReturns(sidetreePeerCfg, nil)
	sidetreeCfgService2.LoadProtocolsReturns(protocolVersions, nil)

	sidetreeCfgService1 := &mocks.SidetreeConfigService{}

	sidetreeCfgProvider := &mocks.SidetreeConfigProvider{}
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

	p.RestartRESTService()
	time.Sleep(20 * time.Millisecond)

	p.Close()
}
