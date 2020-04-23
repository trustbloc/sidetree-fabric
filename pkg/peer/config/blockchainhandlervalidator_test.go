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
	blockchainHandlerCfg                   = `{"BasePath":"/blockchain","MaxTransactionsInResponse":50}`
	blockchainHandlerCfg_NoBasePath        = `{"MaxTransactionsInResponse":50}`
	blockchainHandlerCfg_InvalidBasePath   = `{"BasePath":"blockchain","MaxTransactionsInResponse":50}`
	blockchainHandlerCfg_NoMaxTransactions = `{"BasePath":"/blockchain"}`
)

func TestBlockchainHandlerValidator_Validate(t *testing.T) {
	v := &blockchainHandlerValidator{}

	key := config.NewPeerComponentKey(mspID, peerID, BlockchainHandlerAppName, BlockchainHandlerAppVersion, "/blockchain", BlockchainHandlerComponentVersion)

	t.Run("Valid config -> success", func(t *testing.T) {
		require.NoError(t, v.Validate(config.NewKeyValue(key, config.NewValue(txID, blockchainHandlerCfg, config.FormatJSON))))
	})

	t.Run("No MaxTransactionsInResponse  -> success", func(t *testing.T) {
		require.NoError(t, v.Validate(config.NewKeyValue(key, config.NewValue(txID, blockchainHandlerCfg_NoMaxTransactions, config.FormatJSON))))
	})

	t.Run("Irrelevant config -> success", func(t *testing.T) {
		k1 := config.NewPeerKey(mspID, peerID, "app1", "v1")
		require.NoError(t, v.Validate(config.NewKeyValue(k1, config.NewValue(txID, `{}`, config.FormatJSON))))
	})

	t.Run("Empty config -> error", func(t *testing.T) {
		err := v.Validate(config.NewKeyValue(key, config.NewValue(txID, `{}`, config.FormatJSON)))
		require.Error(t, err)
		require.Contains(t, err.Error(), "field 'BasePath' is required")
	})

	t.Run("No peer ID -> error", func(t *testing.T) {
		k1 := config.NewPeerKey(mspID, "", BlockchainHandlerAppName, BlockchainHandlerAppVersion)
		err := v.Validate(config.NewKeyValue(k1, config.NewValue(txID, `{}`, config.FormatJSON)))
		require.Error(t, err)
		require.Contains(t, err.Error(), "field 'PeerID' required")
	})

	t.Run("Unsupported version -> error", func(t *testing.T) {
		k1 := config.NewPeerKey(mspID, peerID, BlockchainHandlerAppName, "v0.2")
		err := v.Validate(config.NewKeyValue(k1, config.NewValue(txID, `{}`, config.FormatJSON)))
		require.Error(t, err)
		require.Contains(t, err.Error(), "unsupported application version")
	})

	t.Run("Config with no component -> error", func(t *testing.T) {
		k1 := config.NewPeerKey(mspID, peerID, BlockchainHandlerAppName, BlockchainHandlerAppVersion)
		err := v.Validate(config.NewKeyValue(k1, config.NewValue(txID, `{}`, config.FormatJSON)))
		require.Error(t, err)
		require.Contains(t, err.Error(), "empty component name")
	})

	t.Run("Invalid component name -> error", func(t *testing.T) {
		k1 := config.NewPeerComponentKey(mspID, peerID, BlockchainHandlerAppName, BlockchainHandlerAppVersion, "/path", BlockchainHandlerComponentVersion)
		err := v.Validate(config.NewKeyValue(k1, config.NewValue(txID, blockchainHandlerCfg, config.FormatJSON)))
		require.Error(t, err)
		require.Contains(t, err.Error(), "component name must be set to the base path")
	})

	t.Run("Invalid component version -> error", func(t *testing.T) {
		k1 := config.NewPeerComponentKey(mspID, peerID, BlockchainHandlerAppName, BlockchainHandlerAppVersion, "/path", "0.1.2")
		err := v.Validate(config.NewKeyValue(k1, config.NewValue(txID, blockchainHandlerCfg, config.FormatJSON)))
		require.Error(t, err)
		require.Contains(t, err.Error(), "unsupported component version")
	})

	t.Run("Invalid config -> error", func(t *testing.T) {
		err := v.Validate(config.NewKeyValue(key, config.NewValue(txID, `}`, config.FormatJSON)))
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid config")
	})

	t.Run("No BasePath -> error", func(t *testing.T) {
		err := v.Validate(config.NewKeyValue(key, config.NewValue(txID, blockchainHandlerCfg_NoBasePath, config.FormatJSON)))
		require.Error(t, err)
		require.Contains(t, err.Error(), "field 'BasePath' is required")
	})

	t.Run("Invalid BasePath -> error", func(t *testing.T) {
		err := v.Validate(config.NewKeyValue(key, config.NewValue(txID, blockchainHandlerCfg_InvalidBasePath, config.FormatJSON)))
		require.Error(t, err)
		require.Contains(t, err.Error(), "field 'BasePath' must begin with '/'")
	})
}
