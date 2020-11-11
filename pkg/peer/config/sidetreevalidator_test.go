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
	protocolCfg = `{
		"genesisTime":500000,
		"multihashAlgorithm":18,
		"maxOperationSize":2000,
		"maxOperationCount":10,
		"compressionAlgorithm": "GZIP",
		"maxAnchorFileSize": 1000000,
		"maxProofFileSize": 1000000,
		"maxMapFileSize": 1000000,
		"maxChunkFileSize": 10000000,
		"signatureAlgorithms": ["EdDSA", "ES256", "ES256K"],
		"keyAlgorithms": ["Ed25519", "P-256", "secp256k1"],
 		"patches": ["add-public-keys", "remove-public-keys", "add-service-endpoints", "remove-service-endpoints", "ietf-json-patch"]
	}`

	protocolInvalidMulithashAlgoCfg = `{
		"genesisTime":500000,
		"multihashAlgorithm":2777,
		"maxOperationSize":2000,
		"maxOperationCount":10,
		"compressionAlgorithm": "GZIP",
		"maxAnchorFileSize": 1000000,
		"maxProofFileSize": 1000000,
		"maxMapFileSize": 1000000,
		"maxChunkFileSize": 10000000,
		"signatureAlgorithms": ["EdDSA", "ES256", "ES256K"],
		"keyAlgorithms": ["Ed25519", "P-256", "secp256k1"],
 		"patches": ["add-public-keys", "remove-public-keys", "add-service-endpoints", "remove-service-endpoints", "ietf-json-patch"]
	}`

	appCfg = `
batchWriterTimeout: 1s	
chaincodeName: document
collection: docs
`
	appCfgNoCC         = `batchWriterTimeout: 1s`
	appCfgNoCollection = `
batchWriterTimeout: 1s
chaincodeName: document
`
)

func TestSidetreeValidator_Validate(t *testing.T) {
	v := &sidetreeValidator{}

	appKey := config.NewAppKey(GlobalMSPID, "did:sidetree", SidetreeAppVersion)

	t.Run("Irrelevant config -> success", func(t *testing.T) {
		k := config.NewAppKey(mspID, "app1", "v1")
		require.NoError(t, v.Validate(config.NewKeyValue(k, config.NewValue(txID, `{}`, config.FormatJSON))))
	})

	t.Run("Invalid MSP ID -> error", func(t *testing.T) {
		k := config.NewAppKey(mspID, "did:sidetree", SidetreeAppVersion)
		err := v.Validate(config.NewKeyValue(k, config.NewValue(txID, `{}`, config.FormatJSON, sidetreeTag)))
		require.Error(t, err)
		require.Contains(t, err.Error(), "expecting MspID to be set to [general] for Sidetree config")
	})

	t.Run("Unexpected component -> error", func(t *testing.T) {
		k := config.NewComponentKey(GlobalMSPID, "did:sidetree", SidetreeAppVersion, "comp1", "0.4")
		err := v.Validate(config.NewKeyValue(k, config.NewValue(txID, `{}`, config.FormatJSON, sidetreeTag)))
		require.Error(t, err)
		require.Contains(t, err.Error(), "unexpected component")
	})

	t.Run("No chaincodeName -> error", func(t *testing.T) {
		err := v.Validate(config.NewKeyValue(appKey, config.NewValue(txID, appCfgNoCC, config.FormatYAML, sidetreeTag)))
		require.Error(t, err)
		require.Contains(t, err.Error(), "field 'ChaincodeName' is required")
	})

	t.Run("No collection -> error", func(t *testing.T) {
		err := v.Validate(config.NewKeyValue(appKey, config.NewValue(txID, appCfgNoCollection, config.FormatYAML, sidetreeTag)))
		require.Error(t, err)
		require.Contains(t, err.Error(), "field 'Collection' is required")
	})

	t.Run("App config -> success", func(t *testing.T) {
		require.NoError(t, v.Validate(config.NewKeyValue(appKey, config.NewValue(txID, appCfg, config.FormatYAML, sidetreeTag))))
	})

	t.Run("Unsupported app version -> error", func(t *testing.T) {
		k := config.NewAppKey(GlobalMSPID, "did:sidetree", "1.7")
		err := v.Validate(config.NewKeyValue(k, config.NewValue(txID, appCfg, config.FormatYAML, sidetreeTag)))
		require.Error(t, err)
		require.Contains(t, err.Error(), "unsupported application version")
	})

	t.Run("Invalid config -> error", func(t *testing.T) {
		k := config.NewAppKey(GlobalMSPID, "did:sidetree", SidetreeAppVersion)
		err := v.Validate(config.NewKeyValue(k, config.NewValue(txID, `{`, config.FormatJSON, sidetreeTag)))
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid config")
	})

	t.Run("Invalid BatchWriterTimeout -> error", func(t *testing.T) {
		k := config.NewAppKey(GlobalMSPID, "did:sidetree", SidetreeAppVersion)
		err := v.Validate(config.NewKeyValue(k, config.NewValue(txID, `{}`, config.FormatJSON, sidetreeTag)))
		require.Error(t, err)
		require.Contains(t, err.Error(), "field 'BatchWriterTimeout' must contain a value greater than 0")
	})
}

func TestSidetreeValidator_ValidateProtocol(t *testing.T) {
	v := &sidetreeValidator{}

	protocolKey := config.NewComponentKey(GlobalMSPID, "did:sidetree", SidetreeAppVersion, ProtocolComponentName, "0.1")

	t.Run("Protocol -> success", func(t *testing.T) {
		require.NoError(t, v.Validate(config.NewKeyValue(protocolKey, config.NewValue(txID, protocolCfg, config.FormatJSON, sidetreeTag))))
	})

	t.Run("Unsupported version -> success", func(t *testing.T) {
		protocolKey := config.NewComponentKey(GlobalMSPID, "did:sidetree", SidetreeAppVersion, ProtocolComponentName, "0.4")
		err := v.Validate(config.NewKeyValue(protocolKey, config.NewValue(txID, protocolCfg, config.FormatJSON, sidetreeTag)))
		require.Error(t, err)
		require.Contains(t, err.Error(), "protocol version [0.4] not supported")
	})

	t.Run("Invalid config -> error", func(t *testing.T) {
		err := v.Validate(config.NewKeyValue(protocolKey, config.NewValue(txID, `{`, config.FormatJSON, sidetreeTag)))
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid protocol config")
	})

	t.Run("Invalid multihash algorithm -> error", func(t *testing.T) {
		err := v.Validate(config.NewKeyValue(protocolKey, config.NewValue(txID, protocolInvalidMulithashAlgoCfg, config.FormatJSON, sidetreeTag)))
		require.Error(t, err)
		require.Contains(t, err.Error(), "algorithm not supported")
	})
}
