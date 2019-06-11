/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package protocol

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	client, err := New("testdata/protocol.json")
	require.Nil(t, err)
	require.NotNil(t, client)
}

func TestNewError(t *testing.T) {
	client, err := New("testdata/invalid.json")
	require.NotNil(t, err)
	require.Nil(t, client)
	require.Contains(t, err.Error(), "no such file or directory")
}

func TestCurrentProtocol(t *testing.T) {
	client, err := New("testdata/protocol.json")
	require.Nil(t, err)
	require.NotNil(t, client)

	protocol := client.Current()
	require.Equal(t, uint(10000), protocol.MaxOperationsPerBatch)
}
