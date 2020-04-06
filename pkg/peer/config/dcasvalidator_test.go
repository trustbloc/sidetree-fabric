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

func TestDcasValidator_Validate(t *testing.T) {
	v := &dcasValidator{}

	t.Run("Irrelevant config -> success", func(t *testing.T) {
		k := config.NewAppKey(mspID, "app1", "v1")
		require.NoError(t, v.Validate(config.NewKeyValue(k, config.NewValue(txID, `{}`, config.FormatJSON))))
	})

	t.Run("Invalid MSP ID -> error", func(t *testing.T) {
		k := config.NewAppKey(mspID, DCASAppName, DCASAppVersion)
		err := v.Validate(config.NewKeyValue(k, config.NewValue(txID, `{}`, config.FormatJSON)))
		require.Error(t, err)
		require.Contains(t, err.Error(), "expecting MspID to be set to [general] for DCAS config")
	})

	t.Run("Unsupported version -> error", func(t *testing.T) {
		k := config.NewAppKey(GlobalMSPID, DCASAppName, "22")
		err := v.Validate(config.NewKeyValue(k, config.NewValue(txID, `{}`, config.FormatJSON)))
		require.Error(t, err)
		require.Contains(t, err.Error(), "unsupported DCAS config version")
	})

	t.Run("Invalid config -> error", func(t *testing.T) {
		k := config.NewAppKey(GlobalMSPID, DCASAppName, DCASAppVersion)
		err := v.Validate(config.NewKeyValue(k, config.NewValue(txID, `{`, config.FormatJSON)))
		require.Error(t, err)
		require.Contains(t, err.Error(), "unexpected end of JSON input")
	})

	t.Run("Missing ChaincodeName -> error", func(t *testing.T) {
		k := config.NewAppKey(GlobalMSPID, DCASAppName, DCASAppVersion)
		err := v.Validate(config.NewKeyValue(k, config.NewValue(txID, `{}`, config.FormatJSON)))
		require.Error(t, err)
		require.Contains(t, err.Error(), "field 'ChaincodeName' is required")
	})

	t.Run("Missing Collection -> error", func(t *testing.T) {
		k := config.NewAppKey(GlobalMSPID, DCASAppName, DCASAppVersion)
		err := v.Validate(config.NewKeyValue(k, config.NewValue(txID, `{"ChaincodeName":"ccname"}`, config.FormatJSON)))
		require.Error(t, err)
		require.Contains(t, err.Error(), "field 'Collection' is required")
	})

	t.Run("Success", func(t *testing.T) {
		k := config.NewAppKey(GlobalMSPID, DCASAppName, DCASAppVersion)
		err := v.Validate(config.NewKeyValue(k, config.NewValue(txID, `{"ChaincodeName":"ccname","Collection":"coll1"}`, config.FormatJSON)))
		require.NoError(t, err)
	})
}
