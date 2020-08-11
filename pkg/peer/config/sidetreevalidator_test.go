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
	protocolCfg                         = `{"genesisTime":500000,"hashAlgorithmInMultihashCode":18,"maxOperationSize":2000,"maxOperationCount":10}`
	protocolInvalidAlgoCfg              = `{"genesisTime":500000,"hashAlgorithmInMultihashCode":2777,"maxOperationSize":2000,"maxOperationCount":10}`
	protocolInvalidMaxOperationCountCfg = `{"genesisTime":500000,"hashAlgorithmInMultihashCode":18,"maxOperationSize":2000}`
	protocolInvalidMaxOperationSizeCfg  = `{"genesisTime":500000,"hashAlgorithmInMultihashCode":18,"maxOperationCount":10}`
	appCfg                              = `
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

	t.Run("Invalid algorithm -> error", func(t *testing.T) {
		err := v.Validate(config.NewKeyValue(protocolKey, config.NewValue(txID, protocolInvalidAlgoCfg, config.FormatJSON, sidetreeTag)))
		require.Error(t, err)
		require.Contains(t, err.Error(), "algorithm not supported")
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
}
