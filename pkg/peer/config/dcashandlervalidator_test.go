/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package config

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/trustbloc/fabric-peer-ext/pkg/config/ledgerconfig/config"
)

const (
	dcasHandlerCfg                 = `{"BasePath":"/cas","ChaincodeName":"dcascc","Collection":"dcas"}`
	dcasHandlerCfg_NoBasePath      = `{"ChaincodeName":"dcascc","Collection":"dcas"}`
	dcasHandlerCfg_InvalidBasePath = `{"BasePath":"cas","ChaincodeName":"dcascc","Collection":"dcas"}`
	dcasHandlerCfg_NoChaincodeName = `{"BasePath":"/cas","Collection":"dcas"}`
	dcasHandlerCfg_NoCollection    = `{"BasePath":"/cas","ChaincodeName":"dcascc"}`
)

func TestDcasHandlerValidator_Validate(t *testing.T) {
	v := &dcasHandlerValidator{}

	key := config.NewPeerComponentKey(mspID, peerID, DCASHandlerAppName, DCASHandlerAppVersion, "/cas", DCASHandlerComponentVersion)

	t.Run("Valid config -> success", func(t *testing.T) {
		require.NoError(t, v.Validate(config.NewKeyValue(key, config.NewValue(txID, dcasHandlerCfg, config.FormatJSON))))
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
		k1 := config.NewPeerKey(mspID, "", DCASHandlerAppName, DCASHandlerAppVersion)
		err := v.Validate(config.NewKeyValue(k1, config.NewValue(txID, `{}`, config.FormatJSON)))
		require.Error(t, err)
		require.Contains(t, err.Error(), "field PeerID required")
	})

	t.Run("Unsupported version -> error", func(t *testing.T) {
		k1 := config.NewPeerKey(mspID, peerID, DCASHandlerAppName, "v0.2")
		err := v.Validate(config.NewKeyValue(k1, config.NewValue(txID, `{}`, config.FormatJSON)))
		require.Error(t, err)
		require.Contains(t, err.Error(), "unsupported application version")
	})

	t.Run("Config with no component -> error", func(t *testing.T) {
		k1 := config.NewPeerKey(mspID, peerID, DCASHandlerAppName, DCASHandlerAppVersion)
		err := v.Validate(config.NewKeyValue(k1, config.NewValue(txID, `{}`, config.FormatJSON)))
		require.Error(t, err)
		require.Contains(t, err.Error(), "empty component name")
	})

	t.Run("Invalid component name -> error", func(t *testing.T) {
		k1 := config.NewPeerComponentKey(mspID, peerID, DCASHandlerAppName, DCASHandlerAppVersion, "/path", DCASHandlerComponentVersion)
		err := v.Validate(config.NewKeyValue(k1, config.NewValue(txID, dcasHandlerCfg, config.FormatJSON)))
		require.Error(t, err)
		require.Contains(t, err.Error(), "component name must be set to the base path")
	})

	t.Run("Invalid component version -> error", func(t *testing.T) {
		k1 := config.NewPeerComponentKey(mspID, peerID, DCASHandlerAppName, DCASHandlerAppVersion, "/path", "0.1.2")
		err := v.Validate(config.NewKeyValue(k1, config.NewValue(txID, dcasHandlerCfg, config.FormatJSON)))
		require.Error(t, err)
		require.Contains(t, err.Error(), "unsupported component version")
	})

	t.Run("Invalid config -> error", func(t *testing.T) {
		err := v.Validate(config.NewKeyValue(key, config.NewValue(txID, `}`, config.FormatJSON)))
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid config")
	})

	t.Run("No BasePath -> error", func(t *testing.T) {
		err := v.Validate(config.NewKeyValue(key, config.NewValue(txID, dcasHandlerCfg_NoBasePath, config.FormatJSON)))
		require.Error(t, err)
		require.Contains(t, err.Error(), "field 'BasePath' is required")
	})

	t.Run("Invalid BasePath -> error", func(t *testing.T) {
		err := v.Validate(config.NewKeyValue(key, config.NewValue(txID, dcasHandlerCfg_InvalidBasePath, config.FormatJSON)))
		require.Error(t, err)
		require.Contains(t, err.Error(), "field 'BasePath' must begin with '/'")
	})

	t.Run("No ChaincodeName -> error", func(t *testing.T) {
		err := v.Validate(config.NewKeyValue(key, config.NewValue(txID, dcasHandlerCfg_NoChaincodeName, config.FormatJSON)))
		require.Error(t, err)
		require.Contains(t, err.Error(), "field 'ChaincodeName' is required")
	})

	t.Run("No Collection -> error", func(t *testing.T) {
		err := v.Validate(config.NewKeyValue(key, config.NewValue(txID, dcasHandlerCfg_NoCollection, config.FormatJSON)))
		require.Error(t, err)
		require.Contains(t, err.Error(), "field 'Collection' is required")
	})
}
