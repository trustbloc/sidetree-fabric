/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package protocol

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/trustbloc/sidetree-core-go/pkg/api/protocol"
)

func TestNew(t *testing.T) {
	versions := map[string]protocol.Protocol{}
	client := New(versions)
	require.NotNil(t, client)
}

func TestClient_Current(t *testing.T) {
	versions := map[string]protocol.Protocol{
		"1.0": {
			StartingBlockChainTime:       500000,
			HashAlgorithmInMultiHashCode: 18,
			MaxDeltaByteSize:             2000,
			MaxOperationsPerBatch:        10000,
		},
		"0.1": {
			StartingBlockChainTime:       0,
			HashAlgorithmInMultiHashCode: 18,
			MaxDeltaByteSize:             500,
			MaxOperationsPerBatch:        100,
		},
	}

	client := New(versions)
	require.NotNil(t, client)

	protocol := client.Current()
	require.Equal(t, uint(10000), protocol.MaxOperationsPerBatch)
}

func TestClient_Get(t *testing.T) {
	versions := map[string]protocol.Protocol{
		"1.0": {
			StartingBlockChainTime:       500000,
			HashAlgorithmInMultiHashCode: 18,
			MaxDeltaByteSize:             2000,
			MaxOperationsPerBatch:        10000,
		},
		"0.1": {
			StartingBlockChainTime:       10,
			HashAlgorithmInMultiHashCode: 18,
			MaxDeltaByteSize:             500,
			MaxOperationsPerBatch:        100,
		},
	}

	client := New(versions)
	require.NotNil(t, client)

	protocol, err := client.Get(100)
	require.NoError(t, err)
	require.Equal(t, uint(100), protocol.MaxOperationsPerBatch)

	protocol, err = client.Get(500000)
	require.NoError(t, err)
	require.Equal(t, uint(10000), protocol.MaxOperationsPerBatch)

	protocol, err = client.Get(7000000)
	require.NoError(t, err)
	require.Equal(t, uint(10000), protocol.MaxOperationsPerBatch)

	protocol, err = client.Get(5)
	require.Error(t, err)
	require.Equal(t, err.Error(), "protocol parameters are not defined for blockchain time: 5")
}
