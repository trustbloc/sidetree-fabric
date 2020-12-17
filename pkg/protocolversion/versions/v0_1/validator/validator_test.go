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

func TestValidator_Validate(t *testing.T) {
	t.Run("Protocol -> success", func(t *testing.T) {
		p := getProtocol()
		require.NoError(t, Validate(p))
	})

	t.Run("Missing multihash algorithms -> error", func(t *testing.T) {
		protocolInvalidMulithashAlg := getProtocol()
		protocolInvalidMulithashAlg.MultihashAlgorithms = nil

		err := Validate(protocolInvalidMulithashAlg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "field 'MultihashAlgorithms' cannot be empty")
	})

	t.Run("Invalid multihash algorithm -> error", func(t *testing.T) {
		protocolInvalidMulithashAlg := getProtocol()
		protocolInvalidMulithashAlg.MultihashAlgorithms = []uint{2777}

		err := Validate(protocolInvalidMulithashAlg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "error in Sidetree protocol for multihash algorithm(2777): algorithm not supported")
	})

	t.Run("Invalid MaxOperationCount -> error", func(t *testing.T) {
		protocolInvalidMaxOperationCount := getProtocol()
		protocolInvalidMaxOperationCount.MaxOperationCount = 0

		err := Validate(protocolInvalidMaxOperationCount)
		require.Error(t, err)
		require.Contains(t, err.Error(), "field 'MaxOperationCount' must contain a value greater than 0")
	})

	t.Run("Invalid MaxOperationSize -> error", func(t *testing.T) {
		protocolInvalidMaxOperationSize := getProtocol()
		protocolInvalidMaxOperationSize.MaxOperationSize = 0

		err := Validate(protocolInvalidMaxOperationSize)
		require.Error(t, err)
		require.Contains(t, err.Error(), "field 'MaxOperationSize' must contain a value greater than 0")
	})

	t.Run("Invalid MaxDeltaSize -> error", func(t *testing.T) {
		protocolInvalidMaxDeltaSize := getProtocol()
		protocolInvalidMaxDeltaSize.MaxDeltaSize = 0

		err := Validate(protocolInvalidMaxDeltaSize)
		require.Error(t, err)
		require.Contains(t, err.Error(), "field 'MaxDeltaSize' must contain a value greater than 0")
	})

	t.Run("Invalid MaxOperationHashLength -> error", func(t *testing.T) {
		protocolInvalidMaxOperationHashLength := getProtocol()
		protocolInvalidMaxOperationHashLength.MaxOperationHashLength = 0

		err := Validate(protocolInvalidMaxOperationHashLength)
		require.Error(t, err)
		require.Contains(t, err.Error(), "field 'MaxOperationHashLength' must contain a value greater than 0")
	})

	t.Run("Invalid MaxCasURILength -> error", func(t *testing.T) {
		protocolInvalidMaxCasUriLength := getProtocol()
		protocolInvalidMaxCasUriLength.MaxCasURILength = 0

		err := Validate(protocolInvalidMaxCasUriLength)
		require.Error(t, err)
		require.Contains(t, err.Error(), "field 'MaxCasURILength' must contain a value greater than 0")
	})

	t.Run("Invalid MaxCoreIndexFileSize -> error", func(t *testing.T) {
		protocolInvalidMaxCoreIndexSize := getProtocol()
		protocolInvalidMaxCoreIndexSize.MaxCoreIndexFileSize = 0

		err := Validate(protocolInvalidMaxCoreIndexSize)
		require.Error(t, err)
		require.Contains(t, err.Error(), "field 'MaxCoreIndexFileSize' must contain a value greater than 0")
	})

	t.Run("Invalid MaxProofFileSize -> error", func(t *testing.T) {
		protocolInvalidMaxProofSize := getProtocol()
		protocolInvalidMaxProofSize.MaxProofFileSize = 0

		err := Validate(protocolInvalidMaxProofSize)
		require.Error(t, err)
		require.Contains(t, err.Error(), "field 'MaxProofFileSize' must contain a value greater than 0")
	})

	t.Run("Invalid MaxProvisionalIndexFileSize -> error", func(t *testing.T) {
		protocolInvalidMaxProvisionalIndexSize := getProtocol()
		protocolInvalidMaxProvisionalIndexSize.MaxProvisionalIndexFileSize = 0

		err := Validate(protocolInvalidMaxProvisionalIndexSize)
		require.Error(t, err)
		require.Contains(t, err.Error(), "field 'MaxProvisionalIndexFileSize' must contain a value greater than 0")
	})

	t.Run("Invalid MaxChunkFileSize -> error", func(t *testing.T) {
		protocolInvalidMaxChunkSize := getProtocol()
		protocolInvalidMaxChunkSize.MaxChunkFileSize = 0

		err := Validate(protocolInvalidMaxChunkSize)
		require.Error(t, err)
		require.Contains(t, err.Error(), "field 'MaxChunkFileSize' must contain a value greater than 0")
	})

	t.Run("Invalid SignatureAlgorithms -> error", func(t *testing.T) {
		protocolInvalidSignatureAlgorithms := getProtocol()
		protocolInvalidSignatureAlgorithms.SignatureAlgorithms = nil

		err := Validate(protocolInvalidSignatureAlgorithms)
		require.Error(t, err)
		require.Contains(t, err.Error(), "field 'SignatureAlgorithms' cannot be empty")
	})

	t.Run("Invalid KeyAlgorithms -> error", func(t *testing.T) {
		protocolInvalidKeyAlgorithms := getProtocol()
		protocolInvalidKeyAlgorithms.KeyAlgorithms = nil

		err := Validate(protocolInvalidKeyAlgorithms)
		require.Error(t, err)
		require.Contains(t, err.Error(), "field 'KeyAlgorithms' cannot be empty")
	})

	t.Run("Invalid Patches -> error", func(t *testing.T) {
		protocolInvalidPatches := getProtocol()
		protocolInvalidPatches.Patches = nil

		err := Validate(protocolInvalidPatches)
		require.Error(t, err)
		require.Contains(t, err.Error(), "field 'Patches' cannot be empty")
	})

	t.Run("Invalid CompressionAlgorithm -> error", func(t *testing.T) {
		protocolInvalidCompressionAlg := getProtocol()
		protocolInvalidCompressionAlg.CompressionAlgorithm = ""

		err := Validate(protocolInvalidCompressionAlg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "field 'CompressionAlgorithm' cannot be empty")
	})
}

func getProtocol() *protocol.Protocol {
	return &protocol.Protocol{
		GenesisTime:                 500000,
		MultihashAlgorithms:         []uint{18},
		MaxOperationSize:            2000,
		MaxOperationHashLength:      100,
		MaxDeltaSize:                1000,
		MaxCasURILength:             100,
		MaxOperationCount:           10,
		CompressionAlgorithm:        "GZIP",
		MaxCoreIndexFileSize:        1000000,
		MaxProofFileSize:            1000000,
		MaxProvisionalIndexFileSize: 1000000,
		MaxChunkFileSize:            10000000,
		SignatureAlgorithms:         []string{"EdDSA", "ES256", "ES256K"},
		KeyAlgorithms:               []string{"Ed25519", "P-256", "secp256k1"},
		Patches:                     []string{"add-public-keys", "remove-public-keys", "add-service-endpoints", "remove-service-endpoints", "ietf-json-patch"},
	}
}
