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
		r = "read"
		w = "write"
	)

	kv := config.NewKeyValue(
		config.NewAppKey(mspID, "some-app", "v1"),
		config.NewValue(txID, "some config", config.FormatOther),
	)

	t.Run("No tokens -> success", func(t *testing.T) {
		v := newAuthTokenValidator(&peermocks.RestConfig{})
		require.NotNil(t, v)

		cfg := authhandler.Config{}
		require.NoError(t, v.Validate(cfg, kv))
	})

	t.Run("Success", func(t *testing.T) {
		tokenProvider := &peermocks.RestConfig{}
		tokenProvider.SidetreeAPITokenReturns("some-token")

		v := newAuthTokenValidator(tokenProvider)
		require.NotNil(t, v)

		cfg := authhandler.Config{
			ReadTokens:  []string{r, w},
			WriteTokens: []string{w},
		}

		err := v.Validate(cfg, kv)
		require.NoError(t, err)
	})

	t.Run("Read token not defined in config -> error", func(t *testing.T) {
		tokenProvider := &peermocks.RestConfig{}
		tokenProvider.SidetreeAPITokenReturnsOnCall(0, "")
		tokenProvider.SidetreeAPITokenReturnsOnCall(1, "write-token")

		v := newAuthTokenValidator(tokenProvider)
		require.NotNil(t, v)

		cfg := authhandler.Config{
			ReadTokens:  []string{r},
			WriteTokens: []string{w},
		}

		err := v.Validate(cfg, kv)
		require.Error(t, err)
		require.Contains(t, err.Error(), "token name [read] is not defined")
	})

	t.Run("Write token not defined in config -> error", func(t *testing.T) {
		tokenProvider := &peermocks.RestConfig{}
		tokenProvider.SidetreeAPITokenReturnsOnCall(0, "read-token")
		tokenProvider.SidetreeAPITokenReturnsOnCall(1, "")

		v := newAuthTokenValidator(tokenProvider)
		require.NotNil(t, v)

		cfg := authhandler.Config{
			ReadTokens:  []string{r},
			WriteTokens: []string{w},
		}

		err := v.Validate(cfg, kv)
		require.Error(t, err)
		require.Contains(t, err.Error(), "token name [write] is not defined")
	})
}
