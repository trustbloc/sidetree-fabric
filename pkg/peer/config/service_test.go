/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package config

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	ledgercfg "github.com/trustbloc/fabric-peer-ext/pkg/config/ledgerconfig/config"
	"github.com/trustbloc/sidetree-fabric/pkg/peer/config/mocks"
)

//go:generate counterfeiter -o ./mocks/configserviceprovider.gen.go --fake-name ConfigServiceProvider . configServiceProvider
//go:generate counterfeiter -o ./mocks/configservice.gen.go --fake-name ConfigService github.com/trustbloc/fabric-peer-ext/pkg/config/ledgerconfig/config.Service
//go:generate counterfeiter -o ./mocks/validatorregistry.gen.go --fake-name ValidatorRegistry . validatorRegistry

const (
	channelID = "mychannel"
	mspID     = "Org1MSP"
	peerID    = "peer1.example.com"

	v0_4 = "0.4"
	v0_5 = "0.5"

	didSidetreeNamespace             = "did:sidetree"
	didSidetreeCfgJSON               = `{"batchWriterTimeout":"5s"}`
	didSidetreeProtocol_V0_4_CfgJSON = `{"startingBlockchainTime":200000,"hashAlgorithmInMultihashCode":18,"maxOperationByteSize":2000,"maxOperationsPerBatch":10}`
	didSidetreeProtocol_V0_5_CfgJSON = `{"startingBlockchainTime":500000,"hashAlgorithmInMultihashCode":18,"maxOperationByteSize":10000,"maxOperationsPerBatch":100}`
	peerCfgJson                      = `{"Monitor":{"Period":"5s"},"Rest":{"Host":"0.0.0.0","Port":"48326"},"Namespaces":[{"Namespace":"did:sidetree","BasePath":"/document"},{"Namespace":"did:bloc:trustbloc.dev","BasePath":"/trustbloc.dev/document"}]}`
	fileHandlerCfgJson               = `{"Handlers":[{"BasePath":"/schema","ChaincodeName":"files","Collection":"consortium","IndexNamespace":"file:idx","IndexDocID":"file:idx:1234"},{"BasePath":"/.well-known/trustbloc","ChaincodeName":"files","Collection":"consortium","IndexNamespace":"file:idx","IndexDocID":"file:idx:5678"}]}`
)

func TestNewSidetreeProvider(t *testing.T) {
	configService := &mocks.ConfigService{}

	configProvider := &mocks.ConfigServiceProvider{}
	configProvider.ForChannelReturns(configService)

	validatorRegistry := &mocks.ValidatorRegistry{}

	p := NewSidetreeProvider(configProvider, validatorRegistry)
	require.NotNil(t, p)

	s := p.ForChannel(channelID)
	require.NotNil(t, s)

	t.Run("", func(t *testing.T) {
		results := []*ledgercfg.KeyValue{
			{
				Key: ledgercfg.NewComponentKey(GlobalMSPID, didSidetreeNamespace, "1", ProtocolComponentName, v0_4),
				Value: &ledgercfg.Value{
					TxID:   "tx1",
					Format: "json",
					Config: didSidetreeProtocol_V0_4_CfgJSON,
				},
			},
			{
				Key: ledgercfg.NewComponentKey(GlobalMSPID, didSidetreeNamespace, "1", ProtocolComponentName, v0_5),
				Value: &ledgercfg.Value{
					TxID:   "tx2",
					Format: "json",
					Config: didSidetreeProtocol_V0_5_CfgJSON,
				},
			},
		}

		configService.QueryReturns(results, nil)
		protocols, err := s.LoadProtocols(didSidetreeNamespace)
		require.NoError(t, err)
		require.Len(t, protocols, 2)

		protocol4, ok := protocols[v0_4]
		require.True(t, ok)
		require.Equal(t, uint(200000), protocol4.StartingBlockChainTime)
		require.Equal(t, uint(18), protocol4.HashAlgorithmInMultiHashCode)
		require.Equal(t, uint(2000), protocol4.MaxOperationByteSize)
		require.Equal(t, uint(10), protocol4.MaxOperationsPerBatch)

		protocol5, ok := protocols[v0_5]
		require.True(t, ok)
		require.Equal(t, uint(500000), protocol5.StartingBlockChainTime)
		require.Equal(t, uint(18), protocol5.HashAlgorithmInMultiHashCode)
		require.Equal(t, uint(10000), protocol5.MaxOperationByteSize)
		require.Equal(t, uint(100), protocol5.MaxOperationsPerBatch)
	})

	t.Run("LoadSidetree", func(t *testing.T) {
		cfgValue := &ledgercfg.Value{
			TxID:   "tx1",
			Format: "json",
			Config: "{}",
		}

		configService.GetReturns(cfgValue, nil)
		cfg, err := s.LoadSidetree(didSidetreeNamespace)
		require.EqualError(t, err, "batchWriterTimeout must be greater than 0")

		cfgValue = &ledgercfg.Value{
			TxID:   "tx1",
			Format: "json",
			Config: didSidetreeCfgJSON,
		}
		configService.GetReturns(cfgValue, nil)

		cfg, err = s.LoadSidetree(didSidetreeNamespace)
		require.NoError(t, err)
		require.Equal(t, 5*time.Second, cfg.BatchWriterTimeout)
	})

	t.Run("LoadSidetreePeer", func(t *testing.T) {
		cfgValue := &ledgercfg.Value{
			TxID:   "tx1",
			Format: "json",
			Config: peerCfgJson,
		}

		configService.GetReturns(cfgValue, nil)
		cfg, err := s.LoadSidetreePeer(mspID, peerID)
		require.NoError(t, err)
		require.Len(t, cfg.Namespaces, 2)
		require.Equal(t, 5*time.Second, cfg.Monitor.Period)
	})

	t.Run("LoadProtocols service error", func(t *testing.T) {
		errExpected := errors.New("injected config service error")
		configService.QueryReturns(nil, errExpected)

		_, err := s.LoadProtocols(didSidetreeNamespace)
		require.Error(t, err)
		require.Contains(t, err.Error(), errExpected.Error())
	})

	t.Run("LoadSidetree service error", func(t *testing.T) {
		errExpected := errors.New("injected config service error")
		configService.GetReturns(nil, errExpected)

		_, err := s.LoadSidetree(didSidetreeNamespace)
		require.Error(t, err)
		require.Contains(t, err.Error(), errExpected.Error())
	})

	t.Run("LoadSidetreePeer service error", func(t *testing.T) {
		errExpected := errors.New("injected config service error")
		configService.GetReturns(nil, errExpected)

		_, err := s.LoadSidetreePeer(mspID, peerID)
		require.Error(t, err)
		require.Contains(t, err.Error(), errExpected.Error())
	})

	t.Run("load error", func(t *testing.T) {
		cfgValue := &ledgercfg.Value{}
		configService.GetReturns(cfgValue, nil)

		err := s.(*sidetreeService).load(ledgercfg.NewAppKey(GlobalMSPID, SidetreePeerAppName, SidetreePeerAppVersion), func() {})
		require.Error(t, err)
	})

	t.Run("LoadFileHandlers", func(t *testing.T) {
		cfgValue := &ledgercfg.Value{
			TxID:   "tx1",
			Format: "json",
			Config: fileHandlerCfgJson,
		}

		configService.GetReturns(cfgValue, nil)
		cfg, err := s.LoadFileHandlers(mspID, peerID)
		require.NoError(t, err)
		require.Len(t, cfg, 2)
		require.Equal(t, "/schema", cfg[0].BasePath)
	})
}
