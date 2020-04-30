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
	"github.com/trustbloc/sidetree-fabric/pkg/rest/authhandler"
)

func TestAuthTokenValidator_Validate(t *testing.T) {
	const (
		r  = "read"
		rw = "read_write"
	)

	kv := config.NewKeyValue(
		config.NewAppKey(mspID, "some-app", "v1"),
		config.NewValue(txID, "some config", config.FormatOther),
	)

	tokenProvider := &peermocks.RestConfig{}

	v := newAuthTokenValidator(tokenProvider)
	require.NotNil(t, v)

	t.Run("No tokens -> success", func(t *testing.T) {
		cfg := authhandler.Config{}
		require.NoError(t, v.Validate(cfg, kv))
	})

	t.Run("Success", func(t *testing.T) {
		tokenProvider.SidetreeAPITokenReturns("some-token")

		cfg := authhandler.Config{
			ReadTokens:  []string{r, rw},
			WriteTokens: []string{rw},
		}

		err := v.Validate(cfg, kv)
		require.NoError(t, err)
	})

	t.Run("Tokens not defined in peer config -> error", func(t *testing.T) {
		tokenProvider.SidetreeAPITokenReturns("")

		cfg := authhandler.Config{
			ReadTokens:  []string{r, rw},
			WriteTokens: []string{rw},
		}

		err := v.Validate(cfg, kv)
		require.Error(t, err)
		require.Contains(t, err.Error(), "not defined in peer config")
	})
}
