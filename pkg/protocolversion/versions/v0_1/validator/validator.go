/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package validator

import (
	"crypto"

	"github.com/pkg/errors"

	"github.com/trustbloc/sidetree-core-go/pkg/api/protocol"
	"github.com/trustbloc/sidetree-core-go/pkg/docutil"
	"github.com/trustbloc/sidetree-core-go/pkg/hashing"
)

// Validate validates the parameters on the given protococol
func Validate(p *protocol.Protocol) error {
	if _, err := docutil.GetHash(p.MultihashAlgorithm); err != nil {
		return errors.WithMessagef(err, "error in Sidetree protocol")
	}

	if _, err := hashing.GetHash(crypto.Hash(p.HashAlgorithm), []byte("")); err != nil {
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

	if p.MaxAnchorFileSize == 0 {
		return errors.Errorf(errMsg, "MaxAnchorFileSize")
	}

	if p.MaxMapFileSize == 0 {
		return errors.Errorf(errMsg, "MaxMapFileSize")
	}

	if p.MaxChunkFileSize == 0 {
		return errors.Errorf(errMsg, "MaxChunkFileSize")
	}

	return nil
}
