/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package config

import (
	"fmt"
	"testing"

	viper "github.com/spf13/viper2015"
	"github.com/stretchr/testify/require"
)

const (
	host = "localhost"
	port = 1200
)

func TestPeerConfig(t *testing.T) {
	t.Run("Port not set -> error", func(t *testing.T) {
		viper.Reset()

		cfg := NewPeer()
		require.NotNil(t, cfg)

		_, err := cfg.SidetreeListenURL()
		require.EqualError(t, err, "port is not set for REST service")
	})

	t.Run("Host not set -> success", func(t *testing.T) {
		viper.Reset()
		viper.Set("sidetree.port", port)

		cfg := NewPeer()
		require.NotNil(t, cfg)

		url, err := cfg.SidetreeListenURL()
		require.NoError(t, err)
		require.Equal(t, fmt.Sprintf("0.0.0.0:%d", port), url)
	})

	t.Run("Host set -> success", func(t *testing.T) {
		viper.Reset()
		viper.Set("sidetree.port", port)
		viper.Set("sidetree.host", host)

		cfg := NewPeer()
		require.NotNil(t, cfg)

		url, err := cfg.SidetreeListenURL()
		require.NoError(t, err)
		require.Equal(t, fmt.Sprintf("%s:%d", host, port), url)
	})
}
