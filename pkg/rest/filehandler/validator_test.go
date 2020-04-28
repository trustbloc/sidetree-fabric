/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package filehandler

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/trustbloc/sidetree-core-go/pkg/api/batch"
	"github.com/trustbloc/sidetree-core-go/pkg/document"
	"github.com/trustbloc/sidetree-core-go/pkg/docutil"
	"github.com/trustbloc/sidetree-core-go/pkg/patch"
	"github.com/trustbloc/sidetree-core-go/pkg/restapi/helper"
	"github.com/trustbloc/sidetree-core-go/pkg/restapi/model"
	"github.com/trustbloc/sidetree-core-go/pkg/util/ecsigner"

	"github.com/trustbloc/sidetree-fabric/pkg/mocks"
)

const sha2_256 = 18

func TestDocumentValidator_IsValidOriginalDocument(t *testing.T) {
	v := NewValidator(&mocks.OperationStore{})
	require.NotNil(t, v)

	t.Run("Invalid document", func(t *testing.T) {
		doc := &FileIndexDoc{
			ID: "file:idx:1234",
		}
		docBytes, err := json.Marshal(doc)
		require.NoError(t, err)
		require.EqualError(t, v.IsValidOriginalDocument(docBytes), "document must NOT have the id property")
	})

	t.Run("No basePath", func(t *testing.T) {
		doc := &FileIndexDoc{
			UniqueSuffix: "1234",
		}
		docBytes, err := json.Marshal(doc)
		require.NoError(t, err)
		require.EqualError(t, v.IsValidOriginalDocument(docBytes), "missing base path")
	})

	t.Run("No file name", func(t *testing.T) {
		doc := &FileIndexDoc{
			UniqueSuffix: "1234",
			FileIndex: FileIndex{
				BasePath: "/schema",
				Mappings: map[string]string{
					"": "",
				},
			},
		}
		docBytes, err := json.Marshal(doc)
		require.NoError(t, err)
		require.EqualError(t, v.IsValidOriginalDocument(docBytes), "missing file name in mapping")
	})

	t.Run("No file ID", func(t *testing.T) {
		doc := &FileIndexDoc{
			UniqueSuffix: "1234",
			FileIndex: FileIndex{
				BasePath: "/schema",
				Mappings: map[string]string{
					"schema1.json": "",
				},
			},
		}
		docBytes, err := json.Marshal(doc)
		require.NoError(t, err)
		require.EqualError(t, v.IsValidOriginalDocument(docBytes), "missing ID for file name [schema1.json]")
	})

	t.Run("Success", func(t *testing.T) {
		doc := &FileIndexDoc{
			UniqueSuffix: "1234",
			FileIndex: FileIndex{
				BasePath: "/schema",
				Mappings: map[string]string{
					"schema1.json": "1234567",
				},
			},
		}
		docBytes, err := json.Marshal(doc)
		require.NoError(t, err)
		require.NoError(t, v.IsValidOriginalDocument(docBytes))
	})
}

func TestDocumentValidator_IsValidPayload(t *testing.T) {
	s := &mocks.OperationStore{}
	s.GetReturns([]*batch.Operation{{}}, nil)

	v := NewValidator(s)
	require.NotNil(t, v)

	t.Run("Invalid document", func(t *testing.T) {
		require.EqualError(t, v.IsValidPayload([]byte("{}")), "missing unique suffix")
	})

	t.Run("Unmarshal operation error", func(t *testing.T) {
		errExpected := errors.New("injected unmarshal op error")

		restore := unmarshalUpdateOperation
		unmarshalUpdateOperation = func([]byte) (s string, data *model.DeltaModel, err error) { return "", nil, errExpected }
		defer func() { unmarshalUpdateOperation = restore }()

		req, err := getUpdateRequest(`[{"op": "add", "path": "/fileIndex/mappings/schema1.json", "value": "ew3e23w3"}]`)
		require.NoError(t, err)
		require.EqualError(t, v.IsValidPayload(req), errExpected.Error())
	})

	t.Run("No path in JSON patch", func(t *testing.T) {
		req, err := getUpdateRequest(`[{"op": "replace"}]`)
		require.NoError(t, err)
		require.EqualError(t, v.IsValidPayload(req), "invalid JSON patch: path not found")
	})

	t.Run("Unmarshal path error", func(t *testing.T) {
		errExpected := errors.New("injected unmarshal JSON error")

		restore := jsonUnmarshal
		jsonUnmarshal = func(bytes []byte, obj interface{}) error { return errExpected }
		defer func() { jsonUnmarshal = restore }()

		req, err := getUpdateRequest(`[{"op": "add", "path": "/fileIndex/mappings/schema1.json", "value": "ew3e23w3"}]`)
		require.NoError(t, err)
		require.EqualError(t, v.IsValidPayload(req), "invalid JSON patch: invalid path")
	})

	t.Run("Attempt to modify non-mappings section", func(t *testing.T) {
		req, err := getUpdateRequest(`[{"op": "replace", "path": "/fileIndex", "value": "ew3e23w3"}]`)
		require.NoError(t, err)
		require.EqualError(t, v.IsValidPayload(req), "invalid JSON patch: only the mappings section of a file index document may be modified")
	})

	t.Run("Success", func(t *testing.T) {
		req, err := getUpdateRequest(`[{"op": "add", "path": "/fileIndex/mappings/schema1.json", "value": "ew3e23w3"}]`)
		require.NoError(t, err)
		require.NoError(t, v.IsValidPayload(req))
	})
}

func TestDocumentValidator_TransformDocument(t *testing.T) {
	v := NewValidator(&mocks.OperationStore{})
	require.NotNil(t, v)

	t.Run("empty document - success", func(t *testing.T) {
		doc := make(document.Document)
		transformed, err := v.TransformDocument(doc)
		require.NoError(t, err)
		require.Equal(t, doc, transformed.Document)
	})

	t.Run("document with operation keys", func(t *testing.T) {
		doc, err := document.FromBytes([]byte(validDocWithOpsKeysOnly))
		require.NoError(t, err)

		result, err := v.TransformDocument(doc)
		require.NoError(t, err)
		require.Equal(t, 1, len(result.MethodMetadata.OperationPublicKeys))

		jsonTransformed, err := json.Marshal(result.Document)
		require.NoError(t, err)
		didDoc, err := document.DidDocumentFromBytes(jsonTransformed)
		require.NoError(t, err)
		require.Equal(t, 0, len(didDoc.PublicKeys()))
	})

	t.Run("document with mixed operation and general keys", func(t *testing.T) {
		// most likely this scenario will not be used
		doc, err := document.FromBytes([]byte(validDocWithMixedKeys))
		require.NoError(t, err)

		result, err := v.TransformDocument(doc)
		require.NoError(t, err)
		require.Equal(t, 1, len(result.MethodMetadata.OperationPublicKeys))

		jsonTransformed, err := json.Marshal(result.Document)
		require.NoError(t, err)
		didDoc, err := document.DidDocumentFromBytes(jsonTransformed)
		require.NoError(t, err)
		require.Equal(t, 1, len(didDoc.PublicKeys()))
	})
}

func TestUnmarshalUpdateOperation(t *testing.T) {
	t.Run("Invalid payload", func(t *testing.T) {
		suffix, op, err := unmarshalUpdateOperation([]byte("{"))
		require.EqualError(t, err, "invalid update request")
		require.Empty(t, suffix)
		require.Nil(t, op)
	})

	t.Run("Invalid base64 encoding", func(t *testing.T) {
		req := &model.UpdateRequest{
			Delta: `{"%sde3":"-+"}`,
		}

		reqBytes, err := json.Marshal(req)
		require.NoError(t, err)

		suffix, op, err := unmarshalUpdateOperation(reqBytes)
		require.EqualError(t, err, "invalid patch data")
		require.Empty(t, suffix)
		require.Nil(t, op)
	})

	t.Run("Invalid patch data", func(t *testing.T) {
		encodedOp := docutil.EncodeToString([]byte("{"))

		req := &model.UpdateRequest{
			Delta: encodedOp,
		}

		reqBytes, err := json.Marshal(req)
		require.NoError(t, err)

		suffix, patchData, err := unmarshalUpdateOperation(reqBytes)
		require.EqualError(t, err, "invalid patch data")
		require.Empty(t, suffix)
		require.Nil(t, patchData)
	})
}

func TestValidatePatch(t *testing.T) {
	t.Run("not ietf-json-patch", func(t *testing.T) {
		p, err := patch.NewAddPublicKeysPatch("{}")
		require.NoError(t, err)

		err = validatePatch(p)
		require.Error(t, err)
		require.Contains(t, err.Error(), "patch action 'add-public-keys' not supported")
	})
	t.Run("missing patch value", func(t *testing.T) {
		p, err := patch.NewJSONPatch(`[{"op": "add", "path": "path", "value": "value"}]`)
		require.NoError(t, err)
		p["patches"] = ""

		err = validatePatch(p)
		require.Error(t, err)
		require.Contains(t, err.Error(), "missing patches string value")
	})
	t.Run("invalid json patch", func(t *testing.T) {
		p, err := patch.NewJSONPatch(`[{"op": "add", "path": "path", "value": "value"}]`)
		require.NoError(t, err)
		p["patches"] = "invalid"

		err = validatePatch(p)
		require.Error(t, err)
		require.Contains(t, err.Error(), "cannot unmarshal string into Go value of type jsonpatch.Patch")
	})
}

func getUpdateRequest(patches string) ([]byte, error) {
	updatePatch, err := newJSONPatch(patches)
	if err != nil {
		return nil, err
	}

	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}

	return helper.NewUpdateRequest(
		&helper.UpdateRequestInfo{
			DidSuffix:     "1234",
			Patch:         updatePatch,
			MultihashCode: sha2_256,
			Signer:        ecsigner.New(privateKey, "ES256", "update-key"),
		})
}

// newJSONPatch creates new generic update patch without validation
func newJSONPatch(patches string) (patch.Patch, error) {
	var generic []interface{}
	err := json.Unmarshal([]byte(patches), &generic)
	if err != nil {
		return nil, err
	}

	p := make(patch.Patch)
	p[patch.ActionKey] = patch.JSONPatch
	p[patch.PatchesKey] = generic

	return p, nil
}

const validDocWithOpsKeysOnly = `
{
  "id" : "doc:method:abc",
  "publicKey": [
    {
      "id": "update-key",
      "type": "JwsVerificationKey2020",
      "usage": ["ops"],
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

const validDocWithMixedKeys = `
{
  "id" : "doc:method:abc",
  "publicKey": [
    {
      "id": "update-key",
      "type": "JwsVerificationKey2020",
      "usage": ["ops"],
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
      "usage": ["general"],
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
