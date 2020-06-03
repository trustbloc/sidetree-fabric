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
	discoveryHandlerCfg                 = `{"BasePath":"/discovery","Authorization":{"ReadTokens":["discovery_r"]}}`
	discoveryHandlerCfg_NoBasePath      = `{}`
	discoveryHandlerCfg_InvalidBasePath = `{"BasePath":"discovery"}`
)

func TestDiscoveryHandlerValidator_Validate(t *testing.T) {
	tokenProvider := &peermocks.RestConfig{}
	tokenProvider.SidetreeAPITokenReturns("some-token")

	v := newDiscoveryHandlerValidator(tokenProvider)

	key := config.NewPeerComponentKey(mspID, peerID, DiscoveryHandlerAppName, DiscoveryHandlerAppVersion, "/discovery", DiscoveryHandlerComponentVersion)

	t.Run("Valid config -> success", func(t *testing.T) {
		require.NoError(t, v.Validate(config.NewKeyValue(key, config.NewValue(txID, discoveryHandlerCfg, config.FormatJSON))))
	})

	t.Run("Auth tokens not defined -> error", func(t *testing.T) {
		tokenProvider.SidetreeAPITokenReturns("")
		defer tokenProvider.SidetreeAPITokenReturns("some-token")

		err := v.Validate(config.NewKeyValue(key, config.NewValue(txID, discoveryHandlerCfg, config.FormatJSON)))
		require.Error(t, err)
		require.Contains(t, err.Error(), "not defined in peer config")
	})

	t.Run("Irrelevant config -> success", func(t *testing.T) {
		k1 := config.NewPeerKey(mspID, peerID, "app1", "v1")
		require.NoError(t, v.Validate(config.NewKeyValue(k1, config.NewValue(txID, `{}`, config.FormatJSON))))
	})

	t.Run("Empty config -> success", func(t *testing.T) {
		err := v.Validate(config.NewKeyValue(key, config.NewValue(txID, `{}`, config.FormatJSON)))
		require.Error(t, err)
		require.Contains(t, err.Error(), "field 'BasePath' is required")
	})

	t.Run("No peer ID -> error", func(t *testing.T) {
		k1 := config.NewPeerKey(mspID, "", DiscoveryHandlerAppName, DiscoveryHandlerAppVersion)
		err := v.Validate(config.NewKeyValue(k1, config.NewValue(txID, `{}`, config.FormatJSON)))
		require.Error(t, err)
		require.Contains(t, err.Error(), "field PeerID required")
	})

	t.Run("Unsupported version -> error", func(t *testing.T) {
		k1 := config.NewPeerKey(mspID, peerID, DiscoveryHandlerAppName, "v0.2")
		err := v.Validate(config.NewKeyValue(k1, config.NewValue(txID, `{}`, config.FormatJSON)))
		require.Error(t, err)
		require.Contains(t, err.Error(), "unsupported application version")
	})

	t.Run("Config with no component -> error", func(t *testing.T) {
		k1 := config.NewPeerKey(mspID, peerID, DiscoveryHandlerAppName, DiscoveryHandlerAppVersion)
		err := v.Validate(config.NewKeyValue(k1, config.NewValue(txID, `{}`, config.FormatJSON)))
		require.Error(t, err)
		require.Contains(t, err.Error(), "empty component name")
	})

	t.Run("Invalid component name -> error", func(t *testing.T) {
		k1 := config.NewPeerComponentKey(mspID, peerID, DiscoveryHandlerAppName, DiscoveryHandlerAppVersion, "/path", DiscoveryHandlerComponentVersion)
		err := v.Validate(config.NewKeyValue(k1, config.NewValue(txID, discoveryHandlerCfg, config.FormatJSON)))
		require.Error(t, err)
		require.Contains(t, err.Error(), "component name must be set to the base path")
	})

	t.Run("Invalid component version -> error", func(t *testing.T) {
		k1 := config.NewPeerComponentKey(mspID, peerID, DiscoveryHandlerAppName, DiscoveryHandlerAppVersion, "/path", "0.1.2")
		err := v.Validate(config.NewKeyValue(k1, config.NewValue(txID, discoveryHandlerCfg, config.FormatJSON)))
		require.Error(t, err)
		require.Contains(t, err.Error(), "unsupported component version")
	})

	t.Run("Invalid config -> error", func(t *testing.T) {
		err := v.Validate(config.NewKeyValue(key, config.NewValue(txID, `}`, config.FormatJSON)))
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid config")
	})

	t.Run("No BasePath -> error", func(t *testing.T) {
		err := v.Validate(config.NewKeyValue(key, config.NewValue(txID, discoveryHandlerCfg_NoBasePath, config.FormatJSON)))
		require.Error(t, err)
		require.Contains(t, err.Error(), "field 'BasePath' is required")
	})

	t.Run("Invalid BasePath -> error", func(t *testing.T) {
		err := v.Validate(config.NewKeyValue(key, config.NewValue(txID, discoveryHandlerCfg_InvalidBasePath, config.FormatJSON)))
		require.Error(t, err)
		require.Contains(t, err.Error(), "field 'BasePath' must begin with '/'")
	})
}
