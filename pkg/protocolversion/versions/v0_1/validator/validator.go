/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package validator

import (
	"github.com/pkg/errors"

	"github.com/trustbloc/sidetree-core-go/pkg/api/protocol"
	"github.com/trustbloc/sidetree-core-go/pkg/hashing"
)

// Validate validates the parameters on the given protococol
func Validate(p *protocol.Protocol) error {
	if _, err := hashing.GetHashFromMultihash(p.MultihashAlgorithm); err != nil {
		return errors.WithMessagef(err, "error in Sidetree protocol")
	}

	if p.MaxOperationCount == 0 {
		return errors.Errorf("field 'MaxOperationCount' must contain a value greater than 0")
	}

	if p.MaxOperationSize == 0 {
		return errors.Errorf("field 'MaxOperationSize' must contain a value greater than 0")
	}

	if p.CompressionAlgorithm == "" {
		return errors.Errorf("field 'CompressionAlgorithm' cannot be empty")
	}

	if len(p.SignatureAlgorithms) == 0 {
		return errors.Errorf("field 'SignatureAlgorithms' cannot be empty")
	}

	if len(p.KeyAlgorithms) == 0 {
		return errors.Errorf("field 'KeyAlgorithms' cannot be empty")
	}

	if len(p.Patches) == 0 {
		return errors.Errorf("field 'Patches' cannot be empty")
	}

	return verifyBatchSizesV0(p)
}

func verifyBatchSizesV0(p *protocol.Protocol) error {
	const errMsg = "field '%s' must contain a value greater than 0"

	if p.MaxCoreIndexFileSize == 0 {
		return errors.Errorf(errMsg, "MaxCoreIndexFileSize")
	}

	if p.MaxProvisionalIndexFileSize == 0 {
		return errors.Errorf(errMsg, "MaxProvisionalIndexFileSize")
	}

	if p.MaxChunkFileSize == 0 {
		return errors.Errorf(errMsg, "MaxChunkFileSize")
	}

	if p.MaxProofFileSize == 0 {
		return errors.Errorf(errMsg, "MaxProofFileSize")
	}

	if p.MaxDeltaSize == 0 {
		return errors.Errorf(errMsg, "MaxDeltaSize")
	}

	if p.MaxProofSize == 0 {
		return errors.Errorf(errMsg, "MaxProofSize")
	}

	if p.MaxCasURILength == 0 {
		return errors.Errorf(errMsg, "MaxCasURILength")
	}

	if p.MaxOperationHashLength == 0 {
		return errors.Errorf(errMsg, "MaxOperationHashLength")
	}

	return nil
}
