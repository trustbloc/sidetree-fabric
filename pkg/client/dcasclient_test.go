/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package client

import (
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	channel1 = "channel1"
	channel2 = "channel2"
)

func TestDCASClientProvider(t *testing.T) {
	p := NewDCASProvider()
	require.NotNil(t, p)

	client1 := p.ForChannel(channel1)
	require.NotNil(t, client1)

	client2 := p.ForChannel(channel2)
	require.NotNil(t, client2)
	require.NotEqual(t, client1, client2)
}
