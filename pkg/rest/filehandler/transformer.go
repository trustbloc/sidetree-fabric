/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package filehandler

import (
	"github.com/trustbloc/sidetree-core-go/pkg/dochandler/transformer/doctransformer"
	"github.com/trustbloc/sidetree-core-go/pkg/document"
)

// Transformer validates the file index Sidetree document
type Transformer struct {
	*doctransformer.Transformer
}

// NewTransformer returns a new file index document Transformer
func NewTransformer() *Transformer {
	return &Transformer{
		Transformer: doctransformer.New(),
	}
}

// TransformDocument takes internal representation of document and transforms it to required representation
func (v *Transformer) TransformDocument(doc document.Document) (*document.ResolutionResult, error) {
	resolutionResult := &document.ResolutionResult{
		Document:       doc,
		MethodMetadata: document.MethodMetadata{},
	}

	return resolutionResult, nil
}
