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
	org1Peer1Cfg                = `{"Monitor":{"MetaDataChaincodeName":"document","Period":"3s"},"Namespaces":[{"Namespace":"did:sidetree","BasePath":"/document","Authorization":{"ReadTokens":["did_r","did_w"],"WriteTokens": ["did_w"]}}]}`
	org1Peer1CfgNoMetaDataCC    = `{"Monitor":{"Period":"3s"},"Namespaces":[{"Namespace":"did:sidetree","BasePath":"/document"}]}`
	org1Peer1NoNamespaceCfg     = `{"Namespaces":[{"BasePath":"/document"}]}`
	org1Peer1NoBasePathCfg      = `{"Namespaces":[{"Namespace":"did:sidetree"}]}`
	org1Peer1InvalidBasePathCfg = `{"Namespaces":[{"Namespace":"did:sidetree","BasePath":"document"}]}`
)

func TestSidetreePeerValidator_Validate(t *testing.T) {
	tokenProvider := &peermocks.RestConfig{}
	tokenProvider.SidetreeAPITokenReturns("some-token")

	v := newSidetreePeerValidator(tokenProvider)

	key := config.NewPeerKey(mspID, peerID, SidetreePeerAppName, SidetreePeerAppVersion)

	t.Run("Irrelevant config -> success", func(t *testing.T) {
		k1 := config.NewPeerKey(mspID, peerID, "app1", "v1")
		require.NoError(t, v.Validate(config.NewKeyValue(k1, config.NewValue(txID, `{}`, config.FormatJSON))))
	})

	t.Run("Auth tokens not defined -> error", func(t *testing.T) {
		tokenProvider.SidetreeAPITokenReturns("")
		defer tokenProvider.SidetreeAPITokenReturns("some-token")

		err := v.Validate(config.NewKeyValue(key, config.NewValue(txID, org1Peer1Cfg, config.FormatJSON)))
		require.Error(t, err)
		require.Contains(t, err.Error(), "not defined in peer config")
	})

	t.Run("Empty config -> success", func(t *testing.T) {
		require.NoError(t, v.Validate(config.NewKeyValue(key, config.NewValue(txID, `{}`, config.FormatJSON))))
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

	t.Run("Unsupported version -> error", func(t *testing.T) {
		k1 := config.NewPeerKey(mspID, peerID, SidetreePeerAppName, "v0.2")
		err := v.Validate(config.NewKeyValue(k1, config.NewValue(txID, `{}`, config.FormatJSON)))
		require.Error(t, err)
		require.Contains(t, err.Error(), "unsupported application version")
	})

	t.Run("Config with component -> error", func(t *testing.T) {
		k1 := config.NewPeerComponentKey(mspID, peerID, SidetreePeerAppName, SidetreePeerAppVersion, "comp1", "v1")
		err := v.Validate(config.NewKeyValue(k1, config.NewValue(txID, `{}`, config.FormatJSON)))
		require.Error(t, err)
		require.Contains(t, err.Error(), "unexpected component")
	})

	t.Run("Invalid config -> error", func(t *testing.T) {
		err := v.Validate(config.NewKeyValue(key, config.NewValue(txID, `}`, config.FormatJSON)))
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid config")
	})

	t.Run("No Namespace -> error", func(t *testing.T) {
		err := v.Validate(config.NewKeyValue(key, config.NewValue(txID, org1Peer1NoNamespaceCfg, config.FormatJSON)))
		require.Error(t, err)
		require.Contains(t, err.Error(), "field 'Namespace' is required")
	})

	t.Run("No BasePath -> error", func(t *testing.T) {
		err := v.Validate(config.NewKeyValue(key, config.NewValue(txID, org1Peer1NoBasePathCfg, config.FormatJSON)))
		require.Error(t, err)
		require.Contains(t, err.Error(), "field 'BasePath' is required")
	})

	t.Run("Invalid BasePath -> error", func(t *testing.T) {
		err := v.Validate(config.NewKeyValue(key, config.NewValue(txID, org1Peer1InvalidBasePathCfg, config.FormatJSON)))
		require.Error(t, err)
		require.Contains(t, err.Error(), "field 'BasePath' must begin with '/'")
	})
}
