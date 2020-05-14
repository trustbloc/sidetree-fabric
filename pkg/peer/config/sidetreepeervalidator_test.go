/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package config

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/trustbloc/fabric-peer-ext/pkg/config/ledgerconfig/config"

	peermocks "github.com/trustbloc/sidetree-fabric/pkg/peer/mocks"
)

const (
	txID = "tx1"
)

const (
	org1Peer1Cfg                               = `{"Observer":{"MetaDataChaincodeName":"document","Period":"3s"}}`
	org1Peer1NoPeriodCfg                       = `{"Observer":{"MetaDataChaincodeName":"document"}}`
	org1Peer1CfgNoMetaDataCC                   = `{"Observer":{"Period":"3s"}}`
	org1Peer1SidetreeHandlerCfg                = `{"BasePath":"/sidetree/0.0.1","Namespace":"did:sidetree","Authorization":{"ReadTokens":["did_r","did_w"],"WriteTokens": ["did_w"]}}`
	org1Peer1SidetreeHandlerNoNamespaceCfg     = `{"BasePath":"/sidetree/0.0.1"}`
	org1Peer1SidetreeHandlerNoBasePathCfg      = `{"Namespace":"did:sidetree"}`
	org1Peer1SidetreeHandlerInvalidBasePathCfg = `{"Namespace":"did:sidetree","BasePath":"sidetree/0.0.1"}`
)

func TestSidetreePeerValidator_Validate(t *testing.T) {
	tokenProvider := &peermocks.RestConfig{}
	tokenProvider.SidetreeAPITokenReturns("some-token")

	v := newSidetreePeerValidator(tokenProvider)

	key := config.NewPeerKey(mspID, peerID, SidetreePeerAppName, SidetreePeerAppVersion)
	handler1Key := config.NewPeerComponentKey(mspID, peerID, SidetreePeerAppName, SidetreePeerAppVersion, basePath1, v0_1_3)
	handler1InvalidKey := config.NewPeerComponentKey(mspID, peerID, SidetreePeerAppName, SidetreePeerAppVersion, "sidetree/0.0.1", v0_1_3)

	t.Run("Irrelevant config -> success", func(t *testing.T) {
		k1 := config.NewPeerKey(mspID, peerID, "app1", "v1")
		require.NoError(t, v.Validate(config.NewKeyValue(k1, config.NewValue(txID, `{}`, config.FormatJSON))))
	})

	t.Run("Auth tokens not defined -> error", func(t *testing.T) {
		tokenProvider.SidetreeAPITokenReturns("")
		defer tokenProvider.SidetreeAPITokenReturns("some-token")

		err := v.Validate(config.NewKeyValue(handler1Key, config.NewValue(txID, org1Peer1SidetreeHandlerCfg, config.FormatJSON)))
		require.Error(t, err)
		require.Contains(t, err.Error(), "not defined in peer config")
	})

	t.Run("No period -> success", func(t *testing.T) {
		require.NoError(t, v.Validate(config.NewKeyValue(key, config.NewValue(txID, org1Peer1NoPeriodCfg, config.FormatJSON))))
	})

	t.Run("Config with apps -> success", func(t *testing.T) {
		require.NoError(t, v.Validate(config.NewKeyValue(key, config.NewValue(txID, org1Peer1Cfg, config.FormatJSON))))
	})

	t.Run("No MetaDataChaincodeName -> error", func(t *testing.T) {
		err := v.Validate(config.NewKeyValue(key, config.NewValue(txID, org1Peer1CfgNoMetaDataCC, config.FormatJSON)))
		require.Error(t, err)
		require.Contains(t, err.Error(), "field 'MetaDataChaincodeName' is required")
	})

	t.Run("No peer ID -> error", func(t *testing.T) {
		k1 := config.NewPeerKey(mspID, "", SidetreePeerAppName, SidetreePeerAppVersion)
		err := v.Validate(config.NewKeyValue(k1, config.NewValue(txID, `{}`, config.FormatJSON)))
		require.Error(t, err)
		require.Contains(t, err.Error(), "field PeerID required")
	})

	t.Run("Unsupported app version -> error", func(t *testing.T) {
		k1 := config.NewPeerKey(mspID, peerID, SidetreePeerAppName, "v0.2")
		err := v.Validate(config.NewKeyValue(k1, config.NewValue(txID, `{}`, config.FormatJSON)))
		require.Error(t, err)
		require.Contains(t, err.Error(), "unsupported application version")
	})

	t.Run("Handler config -> success", func(t *testing.T) {
		k1 := config.NewPeerComponentKey(mspID, peerID, SidetreePeerAppName, SidetreePeerAppVersion, basePath1, SidetreeHandlerComponentVersion)
		err := v.Validate(config.NewKeyValue(k1, config.NewValue(txID, org1Peer1SidetreeHandlerCfg, config.FormatJSON)))
		require.NoError(t, err)
	})

	t.Run("Unsupported component version -> success", func(t *testing.T) {
		k1 := config.NewPeerComponentKey(mspID, peerID, SidetreePeerAppName, SidetreePeerAppVersion, basePath1, "v1")
		err := v.Validate(config.NewKeyValue(k1, config.NewValue(txID, org1Peer1SidetreeHandlerCfg, config.FormatJSON)))
		require.Error(t, err)
		require.Contains(t, err.Error(), "unsupported component version")
	})

	t.Run("Invalid config -> error", func(t *testing.T) {
		err := v.Validate(config.NewKeyValue(key, config.NewValue(txID, `}`, config.FormatJSON)))
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid config")
	})

	t.Run("Invalid component config -> error", func(t *testing.T) {
		err := v.Validate(config.NewKeyValue(handler1Key, config.NewValue(txID, `}`, config.FormatJSON)))
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid config")
	})

	t.Run("Invalid component name -> error", func(t *testing.T) {
		err := v.Validate(config.NewKeyValue(handler1InvalidKey, config.NewValue(txID, org1Peer1SidetreeHandlerCfg, config.FormatJSON)))
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid component name")
	})

	t.Run("No namespace -> error", func(t *testing.T) {
		err := v.Validate(config.NewKeyValue(handler1Key, config.NewValue(txID, org1Peer1SidetreeHandlerNoNamespaceCfg, config.FormatJSON)))
		require.Error(t, err)
		require.Contains(t, err.Error(), "field 'Namespace' is required")
	})

	t.Run("No BasePath -> error", func(t *testing.T) {
		err := v.Validate(config.NewKeyValue(handler1Key, config.NewValue(txID, org1Peer1SidetreeHandlerNoBasePathCfg, config.FormatJSON)))
		require.Error(t, err)
		require.Contains(t, err.Error(), "field 'BasePath' is required")
	})

	t.Run("Invalid BasePath -> error", func(t *testing.T) {
		err := v.Validate(config.NewKeyValue(handler1InvalidKey, config.NewValue(txID, org1Peer1SidetreeHandlerInvalidBasePathCfg, config.FormatJSON)))
		require.Error(t, err)
		require.Contains(t, err.Error(), "field 'BasePath' must begin with '/'")
	})
}
