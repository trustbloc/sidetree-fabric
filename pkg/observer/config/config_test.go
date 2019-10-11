/*
Copyright SecureKey Technologies Inc. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {

	channels := []string{"ch1", "ch2"}
	monitorPeriod := 5 * time.Second
	cfg := New(channels, monitorPeriod)

	require.Equal(t, channels, cfg.GetChannels())
	require.Equal(t, monitorPeriod, cfg.GetMonitorPeriod())
}
