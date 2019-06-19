/*
Copyright SecureKey Technologies Inc. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {

	channels := []string{"ch1", "ch2"}
	cfg := New(channels)

	require.Equal(t, channels, cfg.GetChannels())

}
