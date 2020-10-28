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
		doc, err := document.FromBytes([]byte(validDoc))
		require.NoError(t, err)

		result, err := v.TransformDocument(doc)
		require.NoError(t, err)

		jsonTransformed, err := json.Marshal(result.Document)
		require.NoError(t, err)
		didDoc, err := document.DidDocumentFromBytes(jsonTransformed)
		require.NoError(t, err)
		require.Equal(t, 0, len(didDoc.PublicKeys()))
	})

}

const validDoc = `
{
  "id" : "doc:method:abc",
  "other": [
    {
      "name": "name"
    }
  ]
}`
