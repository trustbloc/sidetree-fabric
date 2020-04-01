/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package filehandler

import (
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
		unmarshalUpdateOperation = func([]byte) (s string, data *model.PatchDataModel, err error) { return "", nil, errExpected }
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

	doc := make(document.Document)
	transformed, err := v.TransformDocument(doc)
	require.NoError(t, err)
	require.Equal(t, doc, transformed)
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
			PatchData: `{"%sde3":"-+"}`,
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
			PatchData: encodedOp,
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
		p := patch.NewReplacePatch("doc")
		err := validatePatch(p)
		require.Error(t, err)
		require.Contains(t, err.Error(), "patch action 'replace' not supported")
	})
	t.Run("missing patch value", func(t *testing.T) {
		p := patch.NewJSONPatch("")
		err := validatePatch(p)
		require.Error(t, err)
		require.Contains(t, err.Error(), "missing patches string value")
	})
	t.Run("invalid json patch", func(t *testing.T) {
		p := patch.NewJSONPatch("invalid")
		err := validatePatch(p)
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid character 'i'")
	})
}

func getUpdateRequest(patch string) ([]byte, error) {
	return helper.NewUpdateRequest(
		&helper.UpdateRequestInfo{
			DidUniqueSuffix: "1234",
			Patch:           patch,
			MultihashCode:   sha2_256,
		})
}
