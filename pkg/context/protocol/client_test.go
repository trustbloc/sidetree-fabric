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

func TestCurrentProtocol(t *testing.T) {
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
