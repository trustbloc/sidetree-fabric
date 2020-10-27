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

	processKeys(doc)

	return resolutionResult, nil
}

// generic documents will most likely only contain operation keys
// operation keys are not part of external document but resolution result
func processKeys(internal document.Document) {
	var pubKeys []document.PublicKey

	for _, pk := range internal.PublicKeys() {
		externalPK := make(document.PublicKey)
		externalPK[document.IDProperty] = internal.ID() + "#" + pk.ID()
		externalPK[document.TypeProperty] = pk.Type()
		externalPK[document.ControllerProperty] = internal[document.IDProperty]
		externalPK[document.PublicKeyJwkProperty] = pk.PublicKeyJwk()

		delete(pk, document.PurposesProperty)

		pubKeys = append(pubKeys, externalPK)
	}

	if len(pubKeys) > 0 {
		internal[document.PublicKeyProperty] = pubKeys
	} else {
		delete(internal, document.PublicKeyProperty)
	}
}
