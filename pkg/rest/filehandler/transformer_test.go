/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package filehandler

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/trustbloc/sidetree-core-go/pkg/document"
)

func TestDocumentTransformer_TransformDocument(t *testing.T) {
	v := NewTransformer()
	require.NotNil(t, v)

	t.Run("empty document - success", func(t *testing.T) {
		doc := make(document.Document)
		transformed, err := v.TransformDocument(doc)
		require.NoError(t, err)
		require.Equal(t, doc, transformed.Document)
	})

	t.Run("document with no keys", func(t *testing.T) {
		doc, err := document.FromBytes([]byte(validDocNoKeys))
		require.NoError(t, err)

		result, err := v.TransformDocument(doc)
		require.NoError(t, err)

		jsonTransformed, err := json.Marshal(result.Document)
		require.NoError(t, err)
		didDoc, err := document.DidDocumentFromBytes(jsonTransformed)
		require.NoError(t, err)
		require.Equal(t, 0, len(didDoc.PublicKeys()))
	})

	t.Run("document with two general keys", func(t *testing.T) {
		// most likely this scenario will not be used
		doc, err := document.FromBytes([]byte(validDocWithKeys))
		require.NoError(t, err)

		result, err := v.TransformDocument(doc)
		require.NoError(t, err)

		jsonTransformed, err := json.Marshal(result.Document)
		require.NoError(t, err)
		didDoc, err := document.DidDocumentFromBytes(jsonTransformed)
		require.NoError(t, err)
		require.Equal(t, 2, len(didDoc.PublicKeys()))
	})
}

const validDocNoKeys = `
{
  "id" : "doc:method:abc",
  "other": [
    {
      "name": "name"
    }
  ]
}`

// TODO: Revisit if keys are needed for generic documents
const validDocWithKeys = `
{
  "id" : "doc:method:abc",
  "publicKey": [
    {
      "id": "auth-key",
      "type": "JwsVerificationKey2020",
      "purpose": ["general"],
      "jwk": {
        "kty": "EC",
        "crv": "P-256K",
        "x": "PUymIqdtF_qxaAqPABSw-C-owT1KYYQbsMKFM-L9fJA",
        "y": "nM84jDHCMOTGTh_ZdHq4dBBdo4Z5PkEOW9jA8z8IsGc"
      }
    },
    {
      "id": "general-key",
      "type": "JwsVerificationKey2020",
      "purpose": ["general"],
      "jwk": {
        "kty": "EC",
        "crv": "P-256K",
        "x": "PUymIqdtF_qxaAqPABSw-C-owT1KYYQbsMKFM-L9fJA",
        "y": "nM84jDHCMOTGTh_ZdHq4dBBdo4Z5PkEOW9jA8z8IsGc"
      }
    }
  ],
  "other": [
    {
      "name": "name"
    }
  ]
}`
