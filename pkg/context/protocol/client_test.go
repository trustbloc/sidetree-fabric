/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package protocol

import (
	"fmt"
	"testing"

	cb "github.com/hyperledger/fabric-protos-go/common"
	"github.com/stretchr/testify/require"
	"github.com/trustbloc/fabric-peer-ext/pkg/mocks"
	"github.com/trustbloc/sidetree-core-go/pkg/api/protocol"
	coremocks "github.com/trustbloc/sidetree-core-go/pkg/mocks"
)

func TestNew(t *testing.T) {
	client := New(nil, &mocks.Ledger{})
	require.NotNil(t, client)
}

func TestClient_Current(t *testing.T) {
	v1_0 := &coremocks.ProtocolVersion{}
	v1_0.ProtocolReturns(protocol.Protocol{
		GenesisTime:                  500000,
		HashAlgorithmInMultiHashCode: 18,
		MaxOperationSize:             2000,
		MaxOperationCount:            10000,
	})

	v0_1 := &coremocks.ProtocolVersion{}
	v0_1.ProtocolReturns(protocol.Protocol{
		GenesisTime:                  0,
		HashAlgorithmInMultiHashCode: 18,
		MaxOperationSize:             500,
		MaxOperationCount:            100,
	})

	versions := []protocol.Version{v1_0, v0_1}

	l := &mocks.Ledger{
		BlockchainInfo: &cb.BlockchainInfo{Height: 10001},
	}

	client := New(versions, l)
	require.NotNil(t, client)

	p, err := client.Current()
	require.NoError(t, err)
	require.Equal(t, uint(100), p.Protocol().MaxOperationCount)

	l.BlockchainInfo = &cb.BlockchainInfo{Height: 500001}

	p, err = client.Current()
	require.NoError(t, err)
	require.Equal(t, uint(10000), p.Protocol().MaxOperationCount)

	l.BcInfoError = fmt.Errorf("injected protocol error")
	p, err = client.Current()
	require.EqualError(t, err, l.BcInfoError.Error())
}

func TestClient_Get(t *testing.T) {
	v1_0 := &coremocks.ProtocolVersion{}
	v1_0.VersionReturns("1.0")
	v1_0.ProtocolReturns(protocol.Protocol{
		GenesisTime:                  500000,
		HashAlgorithmInMultiHashCode: 18,
		MaxOperationSize:             2000,
		MaxOperationCount:            10000,
	})

	v0_1 := &coremocks.ProtocolVersion{}
	v0_1.VersionReturns("0.1")
	v0_1.ProtocolReturns(protocol.Protocol{
		GenesisTime:                  10,
		HashAlgorithmInMultiHashCode: 18,
		MaxOperationSize:             500,
		MaxOperationCount:            100,
	})

	versions := []protocol.Version{v1_0, v0_1}

	l := &mocks.Ledger{
		BlockchainInfo: &cb.BlockchainInfo{
			Height: 10001,
		},
	}

	client := New(versions, l)
	require.NotNil(t, client)

	protocol, err := client.Get(100)
	require.NoError(t, err)
	require.Equal(t, uint(100), protocol.Protocol().MaxOperationCount)
	require.Equal(t, "0.1", protocol.Version())

	protocol, err = client.Get(500000)
	require.NoError(t, err)
	require.Equal(t, uint(10000), protocol.Protocol().MaxOperationCount)
	require.Equal(t, "1.0", protocol.Version())

	protocol, err = client.Get(7000000)
	require.NoError(t, err)
	require.Equal(t, uint(10000), protocol.Protocol().MaxOperationCount)

	protocol, err = client.Get(5)
	require.Error(t, err)
	require.Equal(t, err.Error(), "protocol parameters are not defined for blockchain time: 5")
}
