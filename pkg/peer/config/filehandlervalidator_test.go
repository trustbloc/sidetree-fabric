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
	fileHandlerCfg                   = `{"BasePath":"/schema","ChaincodeName":"files","Collection":"consortium","IndexNamespace":"file:idx","IndexDocID":"file:idx:1234","Authorization":{"ReadTokens":["content_r","content_w"],"WriteTokens": ["content_w"]}}`
	fileHandlerCfg_NoBasePath        = `{"ChaincodeName":"files","Collection":"consortium","IndexNamespace":"file:idx","IndexDocID":"file:idx:1234"}`
	fileHandlerCfg_InvalidBasePath   = `{"BasePath":"schema","ChaincodeName":"files","Collection":"consortium","IndexNamespace":"file:idx","IndexDocID":"file:idx:1234"}`
	fileHandlerCfg_NoChaincodeName   = `{"BasePath":"/schema","Collection":"consortium","IndexNamespace":"file:idx","IndexDocID":"file:idx:1234"}`
	fileHandlerCfg_NoCollection      = `{"BasePath":"/schema","ChaincodeName":"files","IndexNamespace":"file:idx","IndexDocID":"file:idx:1234"}`
	fileHandlerCfg_NoIndexNamespace  = `{"BasePath":"/schema","ChaincodeName":"files","Collection":"consortium","IndexDocID":"file:idx:1234"}`
	fileHandlerCfg_NoIndexDocID      = `{"BasePath":"/schema","ChaincodeName":"files","Collection":"consortium","IndexNamespace":"file:idx"}`
	fileHandlerCfg_InvalidIndexDocID = `{"BasePath":"/schema","ChaincodeName":"files","Collection":"consortium","IndexNamespace":"file:idx","IndexDocID":"did:bloc:1234"}`
)

func TestFileHandlerValidator_Validate(t *testing.T) {
	tokenProvider := &peermocks.RestConfig{}
	tokenProvider.SidetreeAPITokenReturns("some-token")

	v := newFileHandlerValidator(tokenProvider)

	key := config.NewPeerComponentKey(mspID, peerID, FileHandlerAppName, FileHandlerAppVersion, "/schema", "1")

	t.Run("Valid config -> success", func(t *testing.T) {
		require.NoError(t, v.Validate(config.NewKeyValue(key, config.NewValue(txID, fileHandlerCfg, config.FormatJSON))))
	})

	t.Run("Auth tokens not defined -> error", func(t *testing.T) {
		tokenProvider.SidetreeAPITokenReturns("")
		defer tokenProvider.SidetreeAPITokenReturns("some-token")

		err := v.Validate(config.NewKeyValue(key, config.NewValue(txID, fileHandlerCfg, config.FormatJSON)))
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

	t.Run("Config with no component -> error", func(t *testing.T) {
		k1 := config.NewPeerKey(mspID, peerID, FileHandlerAppName, FileHandlerAppVersion)
		err := v.Validate(config.NewKeyValue(k1, config.NewValue(txID, `{}`, config.FormatJSON)))
		require.Error(t, err)
		require.Contains(t, err.Error(), "empty component name")
	})

	t.Run("Invalid component name -> error", func(t *testing.T) {
		k1 := config.NewPeerComponentKey(mspID, peerID, FileHandlerAppName, FileHandlerAppVersion, "/path", "1")
		err := v.Validate(config.NewKeyValue(k1, config.NewValue(txID, fileHandlerCfg, config.FormatJSON)))
		require.Error(t, err)
		require.Contains(t, err.Error(), "component name must be set to the base path")
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
