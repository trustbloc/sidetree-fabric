/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package sidetreesvc

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/trustbloc/fabric-peer-ext/pkg/gossip/blockpublisher"
	extroles "github.com/trustbloc/fabric-peer-ext/pkg/roles"

	protocolApi "github.com/trustbloc/sidetree-core-go/pkg/api/protocol"
	"github.com/trustbloc/sidetree-core-go/pkg/batch/opqueue"

	"github.com/trustbloc/sidetree-fabric/pkg/filehandler"
	"github.com/trustbloc/sidetree-fabric/pkg/mocks"
	"github.com/trustbloc/sidetree-fabric/pkg/peer/config"
	peermocks "github.com/trustbloc/sidetree-fabric/pkg/peer/mocks"
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
		BatchWriterTimeout: time.Second,
	}

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

	configSvc := &peermocks.ConfigService{}
	configProvider := &peermocks.ConfigServiceProvider{}
	configProvider.ForChannelReturns(configSvc)

	opQueueProvider := &mocks.OperationQueueProvider{}
	opQueueProvider.CreateReturns(&opqueue.MemQueue{}, nil)

	restConfig := &peermocks.RestConfig{}
	restConfig.SidetreeListenURLReturns("localhost:7721", nil)

	dcasClient := &mocks.DCASClient{}
	dcasProvider := &mocks.DCASClientProvider{}
	dcasProvider.ForChannelReturns(dcasClient, nil)

	providers := &providers{
		ContextProviders: &ContextProviders{
			OperationQueueProvider: opQueueProvider,
			DCASProvider:           dcasProvider,
		},
		PeerConfig:     peerConfig,
		RESTConfig:     restConfig,
		ConfigProvider: configProvider,
		BlockPublisher: blockpublisher.NewProvider(),
	}

	fileHandler1 := filehandler.Config{
		BasePath:       "/path",
		ChaincodeName:  "cc1",
		Collection:     "coll1",
		IndexNamespace: "file:idx",
	}

	sidetreeCfgService2 := &peermocks.SidetreeConfigService{}
	sidetreeCfgService2.LoadSidetreeReturns(sidetreeCfg, nil)
	sidetreeCfgService2.LoadSidetreePeerReturns(sidetreePeerCfg, nil)
	sidetreeCfgService2.LoadProtocolsReturns(protocolVersions, nil)
	sidetreeCfgService2.LoadFileHandlersReturns([]filehandler.Config{fileHandler1}, nil)

	sidetreeCfgService1 := &peermocks.SidetreeConfigService{}

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

	p.RestartRESTService()
	time.Sleep(20 * time.Millisecond)

	p.Close()
}
