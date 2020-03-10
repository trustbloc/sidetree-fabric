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
	fileHandlerCfg                   = `{"Handlers":[{"BasePath":"/schema","ChaincodeName":"files","Collection":"consortium","IndexNamespace":"file:idx","IndexDocID":"file:idx:1234"},{"Description":"Consortium .wellknown files","BasePath":"/.well-known/did-bloc","ChaincodeName":"files","Collection":"consortium","IndexNamespace":"file:idx","IndexDocID":"file:idx:3456"}]}`
	fileHandlerCfg_NoBasePath        = `{"Handlers":[{"ChaincodeName":"files","Collection":"consortium","IndexNamespace":"file:idx","IndexDocID":"file:idx:1234"}]}`
	fileHandlerCfg_InvalidBasePath   = `{"Handlers":[{"BasePath":"schema","ChaincodeName":"files","Collection":"consortium","IndexNamespace":"file:idx","IndexDocID":"file:idx:1234"}]}`
	fileHandlerCfg_NoChaincodeName   = `{"Handlers":[{"BasePath":"/schema","Collection":"consortium","IndexNamespace":"file:idx","IndexDocID":"file:idx:1234"}]}`
	fileHandlerCfg_NoCollection      = `{"Handlers":[{"BasePath":"/schema","ChaincodeName":"files","IndexNamespace":"file:idx","IndexDocID":"file:idx:1234"}]}`
	fileHandlerCfg_NoIndexNamespace  = `{"Handlers":[{"BasePath":"/schema","ChaincodeName":"files","Collection":"consortium","IndexDocID":"file:idx:1234"}]}`
	fileHandlerCfg_NoIndexDocID      = `{"Handlers":[{"BasePath":"/schema","ChaincodeName":"files","Collection":"consortium","IndexNamespace":"file:idx"}]}`
	fileHandlerCfg_InvalidIndexDocID = `{"Handlers":[{"BasePath":"/schema","ChaincodeName":"files","Collection":"consortium","IndexNamespace":"file:idx","IndexDocID":"did:bloc:1234"}]}`
)

func TestFileHandlerValidator_Validate(t *testing.T) {
	v := &fileHandlerValidator{}

	key := config.NewPeerKey(mspID, peerID, FileHandlerAppName, FileHandlerAppVersion)

	t.Run("Valid config -> success", func(t *testing.T) {
		require.NoError(t, v.Validate(config.NewKeyValue(key, config.NewValue(txID, fileHandlerCfg, config.FormatJSON))))
	})

	t.Run("Irrelevant config -> success", func(t *testing.T) {
		k1 := config.NewPeerKey(mspID, peerID, "app1", "v1")
		require.NoError(t, v.Validate(config.NewKeyValue(k1, config.NewValue(txID, `{}`, config.FormatJSON))))
	})

	t.Run("Empty config -> success", func(t *testing.T) {
		err := v.Validate(config.NewKeyValue(key, config.NewValue(txID, `{}`, config.FormatJSON)))
		require.Error(t, err)
		require.Contains(t, err.Error(), "expecting at least one file handler")
	})

	t.Run("No peer ID -> error", func(t *testing.T) {
		k1 := config.NewPeerKey(mspID, "", FileHandlerAppName, FileHandlerAppVersion)
		err := v.Validate(config.NewKeyValue(k1, config.NewValue(txID, `{}`, config.FormatJSON)))
		require.Error(t, err)
		require.Contains(t, err.Error(), "field PeerID required")
	})

	t.Run("Unsupported version -> error", func(t *testing.T) {
		k1 := config.NewPeerKey(mspID, peerID, FileHandlerAppName, "v0.2")
		err := v.Validate(config.NewKeyValue(k1, config.NewValue(txID, `{}`, config.FormatJSON)))
		require.Error(t, err)
		require.Contains(t, err.Error(), "unsupported application version")
	})

	t.Run("Config with component -> error", func(t *testing.T) {
		k1 := config.NewPeerComponentKey(mspID, peerID, FileHandlerAppName, FileHandlerAppVersion, "comp1", "v1")
		err := v.Validate(config.NewKeyValue(k1, config.NewValue(txID, `{}`, config.FormatJSON)))
		require.Error(t, err)
		require.Contains(t, err.Error(), "unexpected component")
	})

	t.Run("Invalid config -> error", func(t *testing.T) {
		err := v.Validate(config.NewKeyValue(key, config.NewValue(txID, `}`, config.FormatJSON)))
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid config")
	})

	t.Run("No BasePath -> error", func(t *testing.T) {
		err := v.Validate(config.NewKeyValue(key, config.NewValue(txID, fileHandlerCfg_NoBasePath, config.FormatJSON)))
		require.Error(t, err)
		require.Contains(t, err.Error(), "field 'BasePath' is required")
	})

	t.Run("Invalid BasePath -> error", func(t *testing.T) {
		err := v.Validate(config.NewKeyValue(key, config.NewValue(txID, fileHandlerCfg_InvalidBasePath, config.FormatJSON)))
		require.Error(t, err)
		require.Contains(t, err.Error(), "field 'BasePath' must begin with '/'")
	})

	t.Run("No ChaincodeName -> error", func(t *testing.T) {
		err := v.Validate(config.NewKeyValue(key, config.NewValue(txID, fileHandlerCfg_NoChaincodeName, config.FormatJSON)))
		require.Error(t, err)
		require.Contains(t, err.Error(), "field 'ChaincodeName' is required")
	})

	t.Run("No Collection -> error", func(t *testing.T) {
		err := v.Validate(config.NewKeyValue(key, config.NewValue(txID, fileHandlerCfg_NoCollection, config.FormatJSON)))
		require.Error(t, err)
		require.Contains(t, err.Error(), "field 'Collection' is required")
	})

	t.Run("No IndexNamespace -> error", func(t *testing.T) {
		err := v.Validate(config.NewKeyValue(key, config.NewValue(txID, fileHandlerCfg_NoIndexNamespace, config.FormatJSON)))
		require.Error(t, err)
		require.Contains(t, err.Error(), "field 'IndexNamespace' is required")
	})

	t.Run("No IndexDocID -> success", func(t *testing.T) {
		require.NoError(t, v.Validate(config.NewKeyValue(key, config.NewValue(txID, fileHandlerCfg_NoIndexDocID, config.FormatJSON))))
	})

	t.Run("Invalid IndexDocID -> error", func(t *testing.T) {
		err := v.Validate(config.NewKeyValue(key, config.NewValue(txID, fileHandlerCfg_InvalidIndexDocID, config.FormatJSON)))
		require.Error(t, err)
		require.Contains(t, err.Error(), "field 'IndexDocID' must begin with 'file:idx'")
	})
}
