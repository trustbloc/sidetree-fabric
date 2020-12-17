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

	"github.com/trustbloc/sidetree-fabric/pkg/config"
	cfgmocks "github.com/trustbloc/sidetree-fabric/pkg/peer/config/mocks"
	"github.com/trustbloc/sidetree-fabric/pkg/peer/mocks"
)

//go:generate counterfeiter -o ./mocks/configserviceprovider.gen.go --fake-name ConfigServiceProvider . configServiceProvider
//go:generate counterfeiter -o ./mocks/configservice.gen.go --fake-name ConfigService github.com/trustbloc/fabric-peer-ext/pkg/config/ledgerconfig/config.Service
//go:generate counterfeiter -o ./mocks/validatorregistry.gen.go --fake-name ValidatorRegistry . validatorRegistry

const (
	channelID = "mychannel"
	mspID     = "Org1MSP"
	peerID    = "peer1.example.com"

	v0_4   = "0.4"
	v0_5   = "0.5"
	v0_1_3 = "0.1.3"
	v0_1_4 = "0.1.4"

	basePath1  = "/sidetree/0.0.1"
	namespace1 = "did:sidetree"
	basePath2  = "/file"
	namespace2 = "file:idx"
	alias1     = "did:domain.com"
	alias2     = "did:alias.com"

	didSidetreeNamespace             = "did:sidetree"
	didSidetreeCfgJSON               = `{"batchWriterTimeout":"5s"}`
	didSidetreeCfgJSONWithMethodCtx  = `{"batchWriterTimeout":"5s","methodContext":["ctx1","ctx2"]}`
	didSidetreeCfgJSONWithBase       = `{"batchWriterTimeout":"5s","enableBase":true}`
	didSidetreeProtocol_V0_4_CfgJSON = `{"genesisTime":200000,"multihashAlgorithms":[18],"maxOperationSize":2000,"maxOperationCount":10}`
	didSidetreeProtocol_V0_5_CfgJSON = `{"genesisTime":500000,"multihashAlgorithms":[18],"maxOperationSize":10000,"maxOperationCount":100}`
	sidetreePeerCfgJson              = `{"Observer":{"Period":"5s"}}`
	sidetreeHandler1CfgJson          = `{"Namespace":"did:sidetree","BasePath":"/sidetree/0.0.1","Aliases":["did:domain.com","did:alias.com"]}`
	sidetreeHandler2CfgJson          = `{"Namespace":"file:idx","BasePath":"/file"}`
	fileHandler1CfgJson              = `{"BasePath":"/schema","ChaincodeName":"files","Collection":"consortium","IndexNamespace":"file:idx","IndexDocID":"file:idx:1234"}`
	fileHandler2CfgJson              = `{"BasePath":"/.well-known/trustbloc","ChaincodeName":"files","Collection":"consortium","IndexNamespace":"file:idx","IndexDocID":"file:idx:5678"}`
	dcasHandler1CfgJson              = `{"BasePath":"/0.1.2/cas","ChaincodeName":"cascc","Collection":"cas1"}`
	dcasHandler2CfgJson              = `{"BasePath":"/0.1.3/cas","ChaincodeName":"cascc","Collection":"cas2"}`
	blockchainHandler1CfgJson        = `{"BasePath":"/0.1.2/blockchain"}`
	blockchainHandler2CfgJson        = `{"BasePath":"/0.1.3/blockchain"}`
	discoveryHandler1CfgJson         = `{"BasePath":"/0.1.2/discovery"}`
	discoveryHandler2CfgJson         = `{"BasePath":"/0.1.3/discovery"}`
	dcasCfgJson                      = `{"ChaincodeName":"cc1","Collection":"dcas"}`
	dcasCfgMissingCCJson             = `{"Collection":"dcas"}`
	dcasCfgMissingCollJson           = `{"ChaincodeName":"cc1"}`
)

func TestNewSidetreeProvider(t *testing.T) {
	configService := &mocks.ConfigService{}

	configProvider := &mocks.ConfigServiceProvider{}
	configProvider.ForChannelReturns(configService)

	validatorRegistry := &cfgmocks.ValidatorRegistry{}

	restCfg := &mocks.RestConfig{}

	p := NewSidetreeProvider(configProvider, validatorRegistry, restCfg)
	require.NotNil(t, p)

	s := p.ForChannel(channelID)
	require.NotNil(t, s)

	t.Run("LoadProtocols", func(t *testing.T) {
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
		require.Equal(t, uint64(200000), protocol4.GenesisTime)
		require.Equal(t, uint(18), protocol4.MultihashAlgorithms[0])
		require.Equal(t, uint(2000), protocol4.MaxOperationSize)
		require.Equal(t, uint(10), protocol4.MaxOperationCount)

		protocol5, ok := protocols[v0_5]
		require.True(t, ok)
		require.Equal(t, uint64(500000), protocol5.GenesisTime)
		require.Equal(t, uint(18), protocol5.MultihashAlgorithms[0])
		require.Equal(t, uint(10000), protocol5.MaxOperationSize)
		require.Equal(t, uint(100), protocol5.MaxOperationCount)
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
		require.Equal(t, 0, len(cfg.MethodContext))
		require.Equal(t, false, cfg.EnableBase)

		cfgValue = &ledgercfg.Value{
			TxID:   "tx1",
			Format: "json",
			Config: didSidetreeCfgJSONWithMethodCtx,
		}
		configService.GetReturns(cfgValue, nil)

		cfg, err = s.LoadSidetree(didSidetreeNamespace)
		require.NoError(t, err)
		require.Equal(t, 5*time.Second, cfg.BatchWriterTimeout)
		require.Equal(t, 2, len(cfg.MethodContext))
		require.Equal(t, "ctx1", cfg.MethodContext[0])
		require.Equal(t, "ctx2", cfg.MethodContext[1])

		cfgValue = &ledgercfg.Value{
			TxID:   "tx1",
			Format: "json",
			Config: didSidetreeCfgJSONWithBase,
		}
		configService.GetReturns(cfgValue, nil)

		cfg, err = s.LoadSidetree(didSidetreeNamespace)
		require.NoError(t, err)
		require.Equal(t, 5*time.Second, cfg.BatchWriterTimeout)
		require.Equal(t, 0, len(cfg.MethodContext))
		require.Equal(t, true, cfg.EnableBase)
	})

	t.Run("LoadSidetreePeer", func(t *testing.T) {
		cfgValue := &ledgercfg.Value{
			TxID:   "tx1",
			Format: "json",
			Config: sidetreePeerCfgJson,
		}

		configService.GetReturns(cfgValue, nil)
		cfg, err := s.LoadSidetreePeer(mspID, peerID)
		require.NoError(t, err)
		require.Equal(t, 5*time.Second, cfg.Observer.Period)
	})

	t.Run("LoadSidetreeHandlers", func(t *testing.T) {
		results := []*ledgercfg.KeyValue{
			{
				Key: ledgercfg.NewPeerKey(mspID, peerID, SidetreePeerAppName, SidetreePeerAppVersion),
				Value: &ledgercfg.Value{
					TxID:   "tx1",
					Format: "json",
					Config: sidetreePeerCfgJson,
				},
			},
			{
				Key: ledgercfg.NewPeerComponentKey(mspID, peerID, SidetreePeerAppName, SidetreePeerAppVersion, basePath1, v0_1_3),
				Value: &ledgercfg.Value{
					TxID:   "tx1",
					Format: "json",
					Config: sidetreeHandler1CfgJson,
				},
			},
			{
				Key: ledgercfg.NewPeerComponentKey(mspID, peerID, SidetreePeerAppName, SidetreePeerAppVersion, basePath2, v0_1_4),
				Value: &ledgercfg.Value{
					TxID:   "tx1",
					Format: "json",
					Config: sidetreeHandler2CfgJson,
				},
			},
		}

		configService.QueryReturns(results, nil)
		handlers, err := s.LoadSidetreeHandlers(mspID, peerID)
		require.NoError(t, err)
		require.Len(t, handlers, 2)

		handler1 := handlers[0]
		require.Equal(t, basePath1, handler1.BasePath)
		require.Equal(t, namespace1, handler1.Namespace)
		require.Equal(t, []string{alias1, alias2}, handler1.Aliases)

		handler2 := handlers[1]
		require.Equal(t, basePath2, handler2.BasePath)
		require.Equal(t, namespace2, handler2.Namespace)
		require.Equal(t, 0, len(handler2.Aliases))
	})

	t.Run("LoadSidetreeHandlers query error -> error", func(t *testing.T) {
		errExpected := errors.New("injected config service error")
		configService.QueryReturns(nil, errExpected)

		handlers, err := s.LoadSidetreeHandlers(mspID, peerID)
		require.Error(t, err)
		require.Contains(t, err.Error(), errExpected.Error())
		require.Nil(t, handlers)
	})

	t.Run("LoadSidetreeHandlers unmarshal error", func(t *testing.T) {
		results := []*ledgercfg.KeyValue{
			{
				Key: ledgercfg.NewPeerComponentKey(mspID, peerID, SidetreePeerAppName, SidetreePeerAppVersion, basePath1, v0_1_3),
				Value: &ledgercfg.Value{
					TxID:   "tx1",
					Format: "json",
					Config: `{`,
				},
			},
		}

		configService.QueryReturns(results, nil)
		handlers, err := s.LoadSidetreeHandlers(mspID, peerID)
		require.Error(t, err)
		require.Contains(t, err.Error(), "error reading config")
		require.Empty(t, handlers)
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

	t.Run("LoadFileHandlers query error", func(t *testing.T) {
		errExpected := errors.New("injected query error")

		configService.QueryReturns(nil, errExpected)

		cfg, err := s.LoadFileHandlers(mspID, peerID)
		require.Error(t, err)
		require.Contains(t, err.Error(), errExpected.Error())
		require.Nil(t, cfg)
	})

	t.Run("LoadFileHandlers unmarshal error", func(t *testing.T) {
		queryResults := []*ledgercfg.KeyValue{
			{
				Value: &ledgercfg.Value{
					TxID:   "tx1",
					Format: "json",
					Config: `{`,
				},
			},
		}

		configService.QueryReturns(queryResults, nil)

		cfg, err := s.LoadFileHandlers(mspID, peerID)
		require.Error(t, err)
		require.Contains(t, err.Error(), "error reading config")
		require.Nil(t, cfg)
	})

	t.Run("LoadFileHandlers -> success", func(t *testing.T) {
		queryResults := []*ledgercfg.KeyValue{
			{
				Value: &ledgercfg.Value{
					TxID:   "tx1",
					Format: "json",
					Config: fileHandler1CfgJson,
				},
			},
			{
				Value: &ledgercfg.Value{
					TxID:   "tx2",
					Format: "json",
					Config: fileHandler2CfgJson,
				},
			},
		}

		configService.QueryReturns(queryResults, nil)

		cfg, err := s.LoadFileHandlers(mspID, peerID)
		require.NoError(t, err)
		require.Len(t, cfg, 2)
		require.Equal(t, "/schema", cfg[0].BasePath)
		require.Equal(t, "/.well-known/trustbloc", cfg[1].BasePath)
	})

	t.Run("LoadDCASHandlers query error", func(t *testing.T) {
		errExpected := errors.New("injected query error")

		configService.QueryReturns(nil, errExpected)

		cfg, err := s.LoadDCASHandlers(mspID, peerID)
		require.Error(t, err)
		require.Contains(t, err.Error(), errExpected.Error())
		require.Nil(t, cfg)
	})

	t.Run("LoadDCASHandlers unmarshal error", func(t *testing.T) {
		queryResults := []*ledgercfg.KeyValue{
			{
				Value: &ledgercfg.Value{
					TxID:   "tx1",
					Format: "json",
					Config: `{`,
				},
			},
		}

		configService.QueryReturns(queryResults, nil)

		cfg, err := s.LoadDCASHandlers(mspID, peerID)
		require.Error(t, err)
		require.Contains(t, err.Error(), "error reading config")
		require.Nil(t, cfg)
	})

	t.Run("LoadDCASHandlers -> success", func(t *testing.T) {
		queryResults := []*ledgercfg.KeyValue{
			{
				Key: &ledgercfg.Key{
					ComponentVersion: "1.0",
				},
				Value: &ledgercfg.Value{
					TxID:   "tx1",
					Format: "json",
					Config: dcasHandler1CfgJson,
				},
			},
			{
				Key: &ledgercfg.Key{
					ComponentVersion: "1.0",
				},
				Value: &ledgercfg.Value{
					TxID:   "tx2",
					Format: "json",
					Config: dcasHandler2CfgJson,
				},
			},
		}

		configService.QueryReturns(queryResults, nil)

		cfg, err := s.LoadDCASHandlers(mspID, peerID)
		require.NoError(t, err)
		require.Len(t, cfg, 2)
		require.Equal(t, "/0.1.2/cas", cfg[0].BasePath)
		require.Equal(t, "/0.1.3/cas", cfg[1].BasePath)
	})

	t.Run("LoadBlockchainHandlers query error", func(t *testing.T) {
		errExpected := errors.New("injected query error")

		configService.QueryReturns(nil, errExpected)

		cfg, err := s.LoadBlockchainHandlers(mspID, peerID)
		require.Error(t, err)
		require.Contains(t, err.Error(), errExpected.Error())
		require.Nil(t, cfg)
	})

	t.Run("LoadBlockchainHandlers unmarshal error", func(t *testing.T) {
		queryResults := []*ledgercfg.KeyValue{
			{
				Value: &ledgercfg.Value{
					TxID:   "tx1",
					Format: "json",
					Config: `{`,
				},
			},
		}

		configService.QueryReturns(queryResults, nil)

		cfg, err := s.LoadBlockchainHandlers(mspID, peerID)
		require.Error(t, err)
		require.Contains(t, err.Error(), "error reading config")
		require.Nil(t, cfg)
	})

	t.Run("LoadBlockchainHandlers -> success", func(t *testing.T) {
		queryResults := []*ledgercfg.KeyValue{
			{
				Key: &ledgercfg.Key{
					ComponentVersion: "1.0",
				},
				Value: &ledgercfg.Value{
					TxID:   "tx1",
					Format: "json",
					Config: blockchainHandler1CfgJson,
				},
			},
			{
				Key: &ledgercfg.Key{
					ComponentVersion: "1.0",
				},
				Value: &ledgercfg.Value{
					TxID:   "tx2",
					Format: "json",
					Config: blockchainHandler2CfgJson,
				},
			},
		}

		configService.QueryReturns(queryResults, nil)

		cfg, err := s.LoadBlockchainHandlers(mspID, peerID)
		require.NoError(t, err)
		require.Len(t, cfg, 2)
		require.Equal(t, "/0.1.2/blockchain", cfg[0].BasePath)
		require.Equal(t, "/0.1.3/blockchain", cfg[1].BasePath)
	})

	t.Run("LoadDiscoveryHandlers query error", func(t *testing.T) {
		errExpected := errors.New("injected query error")

		configService.QueryReturns(nil, errExpected)

		cfg, err := s.LoadDiscoveryHandlers(mspID, peerID)
		require.Error(t, err)
		require.Contains(t, err.Error(), errExpected.Error())
		require.Nil(t, cfg)
	})

	t.Run("LoadDiscoveryHandlers unmarshal error", func(t *testing.T) {
		queryResults := []*ledgercfg.KeyValue{
			{
				Value: &ledgercfg.Value{
					TxID:   "tx1",
					Format: "json",
					Config: `{`,
				},
			},
		}

		configService.QueryReturns(queryResults, nil)

		cfg, err := s.LoadDiscoveryHandlers(mspID, peerID)
		require.Error(t, err)
		require.Contains(t, err.Error(), "error reading config")
		require.Nil(t, cfg)
	})

	t.Run("LoadDiscoveryHandlers -> success", func(t *testing.T) {
		queryResults := []*ledgercfg.KeyValue{
			{
				Key: &ledgercfg.Key{
					ComponentVersion: "1.0",
				},
				Value: &ledgercfg.Value{
					TxID:   "tx1",
					Format: "json",
					Config: discoveryHandler1CfgJson,
				},
			},
			{
				Key: &ledgercfg.Key{
					ComponentVersion: "1.0",
				},
				Value: &ledgercfg.Value{
					TxID:   "tx2",
					Format: "json",
					Config: discoveryHandler2CfgJson,
				},
			},
		}

		configService.QueryReturns(queryResults, nil)

		cfg, err := s.LoadDiscoveryHandlers(mspID, peerID)
		require.NoError(t, err)
		require.Len(t, cfg, 2)
		require.Equal(t, "/0.1.2/discovery", cfg[0].BasePath)
		require.Equal(t, "/0.1.3/discovery", cfg[1].BasePath)
	})

	t.Run("LoadDCAS", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			cfgValue := &ledgercfg.Value{
				TxID:   "tx1",
				Format: "json",
				Config: dcasCfgJson,
			}

			configService.GetReturns(cfgValue, nil)

			cfg, err := s.LoadDCAS()
			require.NoError(t, err)
			require.Equal(t, "cc1", cfg.ChaincodeName)
			require.Equal(t, "dcas", cfg.Collection)
		})

		t.Run("Service error", func(t *testing.T) {
			errExpected := errors.New("injected service error")
			configService.GetReturns(nil, errExpected)

			cfg, err := s.LoadDCAS()
			require.Error(t, err)
			require.Contains(t, err.Error(), errExpected.Error())
			require.Equal(t, config.DCAS{}, cfg)
		})

		t.Run("Missing ChaincodeName", func(t *testing.T) {
			cfgValue := &ledgercfg.Value{
				TxID:   "tx1",
				Format: "json",
				Config: dcasCfgMissingCCJson,
			}

			configService.GetReturns(cfgValue, nil)

			cfg, err := s.LoadDCAS()
			require.EqualError(t, err, "field 'ChaincodeName' is required")
			require.Equal(t, config.DCAS{}, cfg)
		})

		t.Run("Missing ChaincodeName", func(t *testing.T) {
			cfgValue := &ledgercfg.Value{
				TxID:   "tx1",
				Format: "json",
				Config: dcasCfgMissingCollJson,
			}

			configService.GetReturns(cfgValue, nil)

			cfg, err := s.LoadDCAS()
			require.EqualError(t, err, "field 'Collection' is required")
			require.Equal(t, config.DCAS{}, cfg)
		})
	})
}
