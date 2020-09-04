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
		"hashAlgorithmInMultihashCode":18,
		"hashAlgorithm":5,
		"maxOperationSize":2000,
		"maxOperationCount":10,
		"compressionAlgorithm": "GZIP",
		"maxAnchorFileSize": 1000000,
		"maxMapFileSize": 1000000,
		"maxChunkFileSize": 10000000,
		"signatureAlgorithms": ["EdDSA", "ES256", "ES256K"],
		"keyAlgorithms": ["Ed25519", "P-256", "secp256k1"]
	}`

	protocolInvalidMulithashAlgoCfg = `{
		"genesisTime":500000,
		"hashAlgorithmInMultihashCode":2777,
		"hashAlgorithm":5,
		"maxOperationSize":2000,
		"maxOperationCount":10,
		"compressionAlgorithm": "GZIP",
		"maxAnchorFileSize": 1000000,
		"maxMapFileSize": 1000000,
		"maxChunkFileSize": 10000000,
		"signatureAlgorithms": ["EdDSA", "ES256", "ES256K"],
		"keyAlgorithms": ["Ed25519", "P-256", "secp256k1"]
	}`

	protocolInvalidHashAlgoCfg = `{
		"genesisTime":500000,
		"hashAlgorithmInMultihashCode":18,
		"maxOperationSize":2000,
		"maxOperationCount":10,
		"compressionAlgorithm": "GZIP",
		"maxAnchorFileSize": 1000000,
		"maxMapFileSize": 1000000,
		"maxChunkFileSize": 10000000,
		"signatureAlgorithms": ["EdDSA", "ES256", "ES256K"],
		"keyAlgorithms": ["Ed25519", "P-256", "secp256k1"]
	}`

	protocolInvalidMaxOperationCountCfg = `{
		"genesisTime":500000,
		"hashAlgorithmInMultihashCode":18,
		"hashAlgorithm":5,
		"maxOperationSize":2000,
		"compressionAlgorithm": "GZIP",
		"maxAnchorFileSize": 1000000,
		"maxMapFileSize": 1000000,
		"maxChunkFileSize": 10000000,
		"signatureAlgorithms": ["EdDSA", "ES256", "ES256K"],
		"keyAlgorithms": ["Ed25519", "P-256", "secp256k1"]
	}`

	protocolInvalidMaxOperationSizeCfg = `{
		"genesisTime":500000,
		"hashAlgorithmInMultihashCode":18,
		"hashAlgorithm":5,
		"maxOperationCount":10,
		"compressionAlgorithm": "GZIP",
		"maxAnchorFileSize": 1000000,
		"maxMapFileSize": 1000000,
		"maxChunkFileSize": 10000000,
		"signatureAlgorithms": ["EdDSA", "ES256", "ES256K"],
		"keyAlgorithms": ["Ed25519", "P-256", "secp256k1"]
	}`

	protocolInvalidMaxAnchorSizeCfg = `{
		"genesisTime":500000,
		"hashAlgorithmInMultihashCode":18,
		"hashAlgorithm":5,
		"maxOperationSize":2000,
		"maxOperationCount":10,
		"compressionAlgorithm": "GZIP",
		"maxMapFileSize": 1000000,
		"maxChunkFileSize": 10000000,
		"signatureAlgorithms": ["EdDSA", "ES256", "ES256K"],
		"keyAlgorithms": ["Ed25519", "P-256", "secp256k1"]
	}`

	protocolInvalidMaxMapSizeCfg = `{
		"genesisTime":500000,
		"hashAlgorithmInMultihashCode":18,
		"hashAlgorithm":5,
		"maxOperationSize":2000,
		"maxOperationCount":10,
		"compressionAlgorithm": "GZIP",
		"maxAnchorFileSize": 1000000,
		"maxChunkFileSize": 10000000,
		"signatureAlgorithms": ["EdDSA", "ES256", "ES256K"],
		"keyAlgorithms": ["Ed25519", "P-256", "secp256k1"]
	}`

	protocolInvalidMaxChunkSizeCfg = `{
		"genesisTime":500000,
		"hashAlgorithmInMultihashCode":18,
		"hashAlgorithm":5,
		"maxOperationSize":2000,
		"maxOperationCount":10,
		"compressionAlgorithm": "GZIP",
		"maxAnchorFileSize": 1000000,
		"maxMapFileSize": 1000000,
		"signatureAlgorithms": ["EdDSA", "ES256", "ES256K"],
		"keyAlgorithms": ["Ed25519", "P-256", "secp256k1"]
	}`

	protocolInvalidCompressionAlgoCfg = `{
		"genesisTime":500000,
		"hashAlgorithmInMultihashCode":18,
		"hashAlgorithm":5,
		"maxOperationSize":2000,
		"maxOperationCount":10,
		"maxAnchorFileSize": 1000000,
		"maxMapFileSize": 1000000,
		"maxChunkFileSize": 10000000,
		"signatureAlgorithms": ["EdDSA", "ES256", "ES256K"],
		"keyAlgorithms": ["Ed25519", "P-256", "secp256k1"]
	}`

	protocolInvalidSignatureAlgorithmsCfg = `{
		"genesisTime":500000,
		"hashAlgorithmInMultihashCode":18,
		"hashAlgorithm":5,
		"maxOperationSize":2000,
		"maxOperationCount":10,
		"compressionAlgorithm": "GZIP",
		"maxAnchorFileSize": 1000000,
		"maxMapFileSize": 1000000,
		"maxChunkFileSize": 10000000,
		"keyAlgorithms": ["Ed25519", "P-256", "secp256k1"]
	}`

	protocolInvalidKeyAlgorithmsCfg = `{
		"genesisTime":500000,
		"hashAlgorithmInMultihashCode":18,
		"hashAlgorithm":5,
		"maxOperationSize":2000,
		"maxOperationCount":10,
		"compressionAlgorithm": "GZIP",
		"maxAnchorFileSize": 1000000,
		"maxMapFileSize": 1000000,
		"maxChunkFileSize": 10000000,
		"signatureAlgorithms": ["EdDSA", "ES256", "ES256K"]
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

	protocolKey := config.NewComponentKey(GlobalMSPID, "did:sidetree", SidetreeAppVersion, ProtocolComponentName, "0.4")

	t.Run("Protocol -> success", func(t *testing.T) {
		require.NoError(t, v.Validate(config.NewKeyValue(protocolKey, config.NewValue(txID, protocolCfg, config.FormatJSON, sidetreeTag))))
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

	t.Run("Invalid hash algorithm -> error", func(t *testing.T) {
		err := v.Validate(config.NewKeyValue(protocolKey, config.NewValue(txID, protocolInvalidHashAlgoCfg, config.FormatJSON, sidetreeTag)))
		require.Error(t, err)
		require.Contains(t, err.Error(), "hash function not available for: 0")
	})

	t.Run("Invalid MaxOperationCount -> error", func(t *testing.T) {
		err := v.Validate(config.NewKeyValue(protocolKey, config.NewValue(txID, protocolInvalidMaxOperationCountCfg, config.FormatJSON, sidetreeTag)))
		require.Error(t, err)
		require.Contains(t, err.Error(), "field 'MaxOperationCount' must contain a value greater than 0")
	})

	t.Run("Invalid MaxOperationSize -> error", func(t *testing.T) {
		err := v.Validate(config.NewKeyValue(protocolKey, config.NewValue(txID, protocolInvalidMaxOperationSizeCfg, config.FormatJSON, sidetreeTag)))
		require.Error(t, err)
		require.Contains(t, err.Error(), "field 'MaxOperationSize' must contain a value greater than 0")
	})

	t.Run("Invalid MaxAnchorFileSize -> error", func(t *testing.T) {
		err := v.Validate(config.NewKeyValue(protocolKey, config.NewValue(txID, protocolInvalidMaxAnchorSizeCfg, config.FormatJSON, sidetreeTag)))
		require.Error(t, err)
		require.Contains(t, err.Error(), "field 'MaxAnchorFileSize' must contain a value greater than 0")
	})

	t.Run("Invalid MaxMapFileSize -> error", func(t *testing.T) {
		err := v.Validate(config.NewKeyValue(protocolKey, config.NewValue(txID, protocolInvalidMaxMapSizeCfg, config.FormatJSON, sidetreeTag)))
		require.Error(t, err)
		require.Contains(t, err.Error(), "field 'MaxMapFileSize' must contain a value greater than 0")
	})

	t.Run("Invalid MaxChunkFileSize -> error", func(t *testing.T) {
		err := v.Validate(config.NewKeyValue(protocolKey, config.NewValue(txID, protocolInvalidMaxChunkSizeCfg, config.FormatJSON, sidetreeTag)))
		require.Error(t, err)
		require.Contains(t, err.Error(), "field 'MaxChunkFileSize' must contain a value greater than 0")
	})

	t.Run("Invalid SignatureAlgorithms -> error", func(t *testing.T) {
		err := v.Validate(config.NewKeyValue(protocolKey, config.NewValue(txID, protocolInvalidSignatureAlgorithmsCfg, config.FormatJSON, sidetreeTag)))
		require.Error(t, err)
		require.Contains(t, err.Error(), "field 'SignatureAlgorithms' cannot be empty")
	})

	t.Run("Invalid KeyAlgorithms -> error", func(t *testing.T) {
		err := v.Validate(config.NewKeyValue(protocolKey, config.NewValue(txID, protocolInvalidKeyAlgorithmsCfg, config.FormatJSON, sidetreeTag)))
		require.Error(t, err)
		require.Contains(t, err.Error(), "field 'KeyAlgorithms' cannot be empty")
	})

	t.Run("Invalid CompressionAlgorithm -> error", func(t *testing.T) {
		err := v.Validate(config.NewKeyValue(protocolKey, config.NewValue(txID, protocolInvalidCompressionAlgoCfg, config.FormatJSON, sidetreeTag)))
		require.Error(t, err)
		require.Contains(t, err.Error(), "field 'CompressionAlgorithm' cannot be empty")
	})
}
