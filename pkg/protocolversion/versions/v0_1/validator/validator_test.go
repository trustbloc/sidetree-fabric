/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package validator

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/trustbloc/sidetree-core-go/pkg/api/protocol"
)

var (
	p = &protocol.Protocol{
		GenesisTime:          500000,
		MultihashAlgorithm:   18,
		HashAlgorithm:        5,
		MaxOperationSize:     2000,
		MaxOperationCount:    10,
		CompressionAlgorithm: "GZIP",
		MaxAnchorFileSize:    1000000,
		MaxMapFileSize:       1000000,
		MaxChunkFileSize:     10000000,
		SignatureAlgorithms:  []string{"EdDSA", "ES256", "ES256K"},
		KeyAlgorithms:        []string{"Ed25519", "P-256", "secp256k1"},
		Patches:              []string{"add-public-keys", "remove-public-keys", "add-service-endpoints", "remove-service-endpoints", "ietf-json-patch"},
	}

	protocolInvalidPatches = &protocol.Protocol{
		GenesisTime:          500000,
		MultihashAlgorithm:   18,
		HashAlgorithm:        5,
		MaxOperationSize:     2000,
		MaxOperationCount:    10,
		CompressionAlgorithm: "GZIP",
		MaxAnchorFileSize:    1000000,
		MaxMapFileSize:       1000000,
		MaxChunkFileSize:     10000000,
		SignatureAlgorithms:  []string{"EdDSA", "ES256", "ES256K"},
		KeyAlgorithms:        []string{"Ed25519", "P-256", "secp256k1"},
	}

	protocolInvalidMulithashAlgo = &protocol.Protocol{
		GenesisTime:          500000,
		MultihashAlgorithm:   2777,
		HashAlgorithm:        5,
		MaxOperationSize:     2000,
		MaxOperationCount:    10,
		CompressionAlgorithm: "GZIP",
		MaxAnchorFileSize:    1000000,
		MaxMapFileSize:       1000000,
		MaxChunkFileSize:     10000000,
		SignatureAlgorithms:  []string{"EdDSA", "ES256", "ES256K"},
		KeyAlgorithms:        []string{"Ed25519", "P-256", "secp256k1"},
		Patches:              []string{"ietf-json-patch"},
	}

	protocolInvalidHashAlgo = &protocol.Protocol{
		GenesisTime:          500000,
		MultihashAlgorithm:   18,
		MaxOperationSize:     2000,
		MaxOperationCount:    10,
		CompressionAlgorithm: "GZIP",
		MaxAnchorFileSize:    1000000,
		MaxMapFileSize:       1000000,
		MaxChunkFileSize:     10000000,
		SignatureAlgorithms:  []string{"EdDSA", "ES256", "ES256K"},
		KeyAlgorithms:        []string{"Ed25519", "P-256", "secp256k1"},
		Patches:              []string{"ietf-json-patch"},
	}

	protocolInvalidMaxOperationCount = &protocol.Protocol{
		GenesisTime:          500000,
		MultihashAlgorithm:   18,
		HashAlgorithm:        5,
		MaxOperationSize:     2000,
		CompressionAlgorithm: "GZIP",
		MaxAnchorFileSize:    1000000,
		MaxMapFileSize:       1000000,
		MaxChunkFileSize:     10000000,
		SignatureAlgorithms:  []string{"EdDSA", "ES256", "ES256K"},
		KeyAlgorithms:        []string{"Ed25519", "P-256", "secp256k1"},
		Patches:              []string{"ietf-json-patch"},
	}

	protocolInvalidMaxOperationSize = &protocol.Protocol{
		GenesisTime:          500000,
		MultihashAlgorithm:   18,
		HashAlgorithm:        5,
		MaxOperationCount:    10,
		CompressionAlgorithm: "GZIP",
		MaxAnchorFileSize:    1000000,
		MaxMapFileSize:       1000000,
		MaxChunkFileSize:     10000000,
		SignatureAlgorithms:  []string{"EdDSA", "ES256", "ES256K"},
		KeyAlgorithms:        []string{"Ed25519", "P-256", "secp256k1"},
		Patches:              []string{"ietf-json-patch"},
	}

	protocolInvalidMaxAnchorSize = &protocol.Protocol{
		GenesisTime:          500000,
		MultihashAlgorithm:   18,
		HashAlgorithm:        5,
		MaxOperationSize:     2000,
		MaxOperationCount:    10,
		CompressionAlgorithm: "GZIP",
		MaxMapFileSize:       1000000,
		MaxChunkFileSize:     10000000,
		SignatureAlgorithms:  []string{"EdDSA", "ES256", "ES256K"},
		KeyAlgorithms:        []string{"Ed25519", "P-256", "secp256k1"},
		Patches:              []string{"ietf-json-patch"},
	}

	protocolInvalidMaxMapSize = &protocol.Protocol{
		GenesisTime:          500000,
		MultihashAlgorithm:   18,
		HashAlgorithm:        5,
		MaxOperationSize:     2000,
		MaxOperationCount:    10,
		CompressionAlgorithm: "GZIP",
		MaxAnchorFileSize:    1000000,
		MaxChunkFileSize:     10000000,
		SignatureAlgorithms:  []string{"EdDSA", "ES256", "ES256K"},
		KeyAlgorithms:        []string{"Ed25519", "P-256", "secp256k1"},
		Patches:              []string{"ietf-json-patch"},
	}

	protocolInvalidMaxChunkSize = &protocol.Protocol{
		GenesisTime:          500000,
		MultihashAlgorithm:   18,
		HashAlgorithm:        5,
		MaxOperationSize:     2000,
		MaxOperationCount:    10,
		CompressionAlgorithm: "GZIP",
		MaxAnchorFileSize:    1000000,
		MaxMapFileSize:       1000000,
		SignatureAlgorithms:  []string{"EdDSA", "ES256", "ES256K"},
		KeyAlgorithms:        []string{"Ed25519", "P-256", "secp256k1"},
		Patches:              []string{"ietf-json-patch"},
	}

	protocolInvalidCompressionAlgo = &protocol.Protocol{
		GenesisTime:         500000,
		MultihashAlgorithm:  18,
		HashAlgorithm:       5,
		MaxOperationSize:    2000,
		MaxOperationCount:   10,
		MaxAnchorFileSize:   1000000,
		MaxMapFileSize:      1000000,
		MaxChunkFileSize:    10000000,
		SignatureAlgorithms: []string{"EdDSA", "ES256", "ES256K"},
		KeyAlgorithms:       []string{"Ed25519", "P-256", "secp256k1"},
	}

	protocolInvalidSignatureAlgorithms = &protocol.Protocol{
		GenesisTime:          500000,
		MultihashAlgorithm:   18,
		HashAlgorithm:        5,
		MaxOperationSize:     2000,
		MaxOperationCount:    10,
		CompressionAlgorithm: "GZIP",
		MaxAnchorFileSize:    1000000,
		MaxMapFileSize:       1000000,
		MaxChunkFileSize:     10000000,
		KeyAlgorithms:        []string{"Ed25519", "P-256", "secp256k1"},
	}

	protocolInvalidKeyAlgorithms = &protocol.Protocol{
		GenesisTime:          500000,
		MultihashAlgorithm:   18,
		HashAlgorithm:        5,
		MaxOperationSize:     2000,
		MaxOperationCount:    10,
		CompressionAlgorithm: "GZIP",
		MaxAnchorFileSize:    1000000,
		MaxMapFileSize:       1000000,
		MaxChunkFileSize:     10000000,
		SignatureAlgorithms:  []string{"EdDSA", "ES256", "ES256K"},
	}
)

func TestValidator_Validate(t *testing.T) {
	t.Run("Protocol -> success", func(t *testing.T) {
		require.NoError(t, Validate(p))
	})

	t.Run("Invalid multihash algorithm -> error", func(t *testing.T) {
		err := Validate(protocolInvalidMulithashAlgo)
		require.Error(t, err)
		require.Contains(t, err.Error(), "algorithm not supported")
	})

	t.Run("Invalid hash algorithm -> error", func(t *testing.T) {
		err := Validate(protocolInvalidHashAlgo)
		require.Error(t, err)
		require.Contains(t, err.Error(), "hash function not available for: 0")
	})

	t.Run("Invalid MaxOperationCount -> error", func(t *testing.T) {
		err := Validate(protocolInvalidMaxOperationCount)
		require.Error(t, err)
		require.Contains(t, err.Error(), "field 'MaxOperationCount' must contain a value greater than 0")
	})

	t.Run("Invalid MaxOperationSize -> error", func(t *testing.T) {
		err := Validate(protocolInvalidMaxOperationSize)
		require.Error(t, err)
		require.Contains(t, err.Error(), "field 'MaxOperationSize' must contain a value greater than 0")
	})

	t.Run("Invalid MaxAnchorFileSize -> error", func(t *testing.T) {
		err := Validate(protocolInvalidMaxAnchorSize)
		require.Error(t, err)
		require.Contains(t, err.Error(), "field 'MaxAnchorFileSize' must contain a value greater than 0")
	})

	t.Run("Invalid MaxMapFileSize -> error", func(t *testing.T) {
		err := Validate(protocolInvalidMaxMapSize)
		require.Error(t, err)
		require.Contains(t, err.Error(), "field 'MaxMapFileSize' must contain a value greater than 0")
	})

	t.Run("Invalid MaxChunkFileSize -> error", func(t *testing.T) {
		err := Validate(protocolInvalidMaxChunkSize)
		require.Error(t, err)
		require.Contains(t, err.Error(), "field 'MaxChunkFileSize' must contain a value greater than 0")
	})

	t.Run("Invalid SignatureAlgorithms -> error", func(t *testing.T) {
		err := Validate(protocolInvalidSignatureAlgorithms)
		require.Error(t, err)
		require.Contains(t, err.Error(), "field 'SignatureAlgorithms' cannot be empty")
	})

	t.Run("Invalid KeyAlgorithms -> error", func(t *testing.T) {
		err := Validate(protocolInvalidKeyAlgorithms)
		require.Error(t, err)
		require.Contains(t, err.Error(), "field 'KeyAlgorithms' cannot be empty")
	})

	t.Run("Invalid Patches -> error", func(t *testing.T) {
		err := Validate(protocolInvalidPatches)
		require.Error(t, err)
		require.Contains(t, err.Error(), "field 'Patches' cannot be empty")
	})

	t.Run("Invalid CompressionAlgorithm -> error", func(t *testing.T) {
		err := Validate(protocolInvalidCompressionAlgo)
		require.Error(t, err)
		require.Contains(t, err.Error(), "field 'CompressionAlgorithm' cannot be empty")
	})
}
