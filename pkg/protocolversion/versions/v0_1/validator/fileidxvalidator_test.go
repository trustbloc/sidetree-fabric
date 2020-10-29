/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package validator

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/json"
	"errors"

	"testing"

	"github.com/stretchr/testify/require"
	"github.com/trustbloc/sidetree-core-go/pkg/api/operation"
	"github.com/trustbloc/sidetree-core-go/pkg/patch"
	"github.com/trustbloc/sidetree-core-go/pkg/util/ecsigner"
	"github.com/trustbloc/sidetree-core-go/pkg/util/pubkey"
	"github.com/trustbloc/sidetree-core-go/pkg/versions/0_1/client"
	"github.com/trustbloc/sidetree-core-go/pkg/versions/0_1/model"

	"github.com/trustbloc/sidetree-fabric/pkg/mocks"
	"github.com/trustbloc/sidetree-fabric/pkg/rest/filehandler"
)

const sha2_256 = 18

func TestDocumentValidator_IsValidOriginalDocument(t *testing.T) {
	v := NewFileIdxValidator(&mocks.OperationStore{})
	require.NotNil(t, v)

	t.Run("Invalid document", func(t *testing.T) {
		doc := &filehandler.FileIndexDoc{
			ID: "file:idx:1234",
		}
		docBytes, err := json.Marshal(doc)
		require.NoError(t, err)
		require.EqualError(t, v.IsValidOriginalDocument(docBytes), "document must NOT have the id property")
	})

	t.Run("No basePath", func(t *testing.T) {
		doc := &filehandler.FileIndexDoc{
			UniqueSuffix: "1234",
		}
		docBytes, err := json.Marshal(doc)
		require.NoError(t, err)
		require.EqualError(t, v.IsValidOriginalDocument(docBytes), "missing base path")
	})

	t.Run("No file name", func(t *testing.T) {
		doc := &filehandler.FileIndexDoc{
			UniqueSuffix: "1234",
			FileIndex: filehandler.FileIndex{
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
		doc := &filehandler.FileIndexDoc{
			UniqueSuffix: "1234",
			FileIndex: filehandler.FileIndex{
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
		doc := &filehandler.FileIndexDoc{
			UniqueSuffix: "1234",
			FileIndex: filehandler.FileIndex{
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
	s.GetReturns([]*operation.AnchoredOperation{{}}, nil)

	v := NewFileIdxValidator(s)
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

func TestUnmarshalUpdateOperation(t *testing.T) {
	t.Run("Invalid payload", func(t *testing.T) {
		suffix, op, err := unmarshalUpdateOperation([]byte("{"))
		require.EqualError(t, err, "invalid update request")
		require.Empty(t, suffix)
		require.Nil(t, op)
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
	t.Run("missing patch action", func(t *testing.T) {
		err := validatePatch(patch.Patch{})
		require.Error(t, err)
		require.Contains(t, err.Error(), "unable to get patch action")
	})

	t.Run("missing patch value", func(t *testing.T) {
		p, err := patch.NewJSONPatch(`[{"op": "add", "path": "path", "value": "value"}]`)
		require.NoError(t, err)
		delete(p, "patches")

		err = validatePatch(p)
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid patch value: ietf-json-patch patch is missing key: patches")
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

	updatePubKey, err := pubkey.GetPublicKeyJWK(&privateKey.PublicKey)
	if err != nil {
		return nil, err
	}

	return client.NewUpdateRequest(
		&client.UpdateRequestInfo{
			DidSuffix:     "1234",
			Patches:       []patch.Patch{updatePatch},
			MultihashCode: sha2_256,
			UpdateKey:     updatePubKey,
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
