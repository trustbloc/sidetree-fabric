/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package filehandler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/trustbloc/sidetree-core-go/pkg/document"

	"github.com/trustbloc/sidetree-fabric/pkg/mocks"
)

//go:generate counterfeiter -o ../mocks/documentresolver.gen.go --fake-name DocumentResolver . documentResolver

const (
	channelID = "channel1"
	schema1   = "schema1.json"
)

func TestFileRetrieveHandler(t *testing.T) {
	docResolver := &mocks.DocumentResolver{}
	dcasProvider := &mocks.DCASClientProvider{}

	dcasClient := mocks.NewDCASClient()
	dcasProvider.GetDCASClientReturns(dcasClient, nil)

	cfg := Config{
		BasePath:       "/schema",
		ChaincodeName:  "file_cc",
		Collection:     "schemas",
		IndexNamespace: "file:idx",
		IndexDocID:     "file:idx:1234",
	}

	h := NewRetrieveHandler(channelID, cfg, docResolver, dcasProvider)
	require.Equal(t, cfg.BasePath+"/{resourceName}", h.Path())
	require.Equal(t, http.MethodGet, h.Method())
	require.NotNil(t, h.Handler())

	t.Run("Bad request", func(t *testing.T) {
		rw := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/schema", nil)
		h.Handler()(rw, req)
		require.Equal(t, http.StatusBadRequest, rw.Code)
		require.Contains(t, rw.Body.String(), "resource name not provided")
	})

	t.Run("File index not found", func(t *testing.T) {
		docResolver.ResolveDocumentReturns(nil, errors.New("not found"))

		getResourceName = func(req *http.Request) string { return schema1 }
		rw := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/schema/schema1.json", nil)
		h.Handler()(rw, req)
		require.Equal(t, http.StatusNotFound, rw.Code)
		require.Equal(t, "file index document not found", rw.Body.String())
	})

	t.Run("File index was deleted", func(t *testing.T) {
		docMetadata := make(document.Metadata)
		docMetadata[document.DeactivatedProperty] = true

		docResolver.ResolveDocumentReturns(&document.ResolutionResult{DocumentMetadata: docMetadata, Document: make(document.Document)}, nil)

		getResourceName = func(req *http.Request) string { return schema1 }
		rw := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/schema/schema1.json", nil)
		h.Handler()(rw, req)
		require.Equal(t, http.StatusGone, rw.Code)
		require.Equal(t, "document is no longer available", rw.Body.String())
	})

	t.Run("Invalid file index document", func(t *testing.T) {
		docResolver.ResolveDocumentReturns(&document.ResolutionResult{}, nil)

		getResourceName = func(req *http.Request) string { return schema1 }
		rw := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/schema/schema1.json", nil)
		h.Handler()(rw, req)
		require.Equal(t, http.StatusInternalServerError, rw.Code)
		require.Equal(t, serverError, rw.Body.String())
	})

	t.Run("ID not found in index", func(t *testing.T) {
		fileIndexDoc := &FileIndexDoc{}

		doc, err := getDocument(fileIndexDoc)
		require.NoError(t, err)

		docResolver.ResolveDocumentReturns(&document.ResolutionResult{Document: doc}, nil)

		getResourceName = func(req *http.Request) string { return schema1 }
		rw := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/schema/schema1.json", nil)
		h.Handler()(rw, req)
		require.Equal(t, http.StatusInternalServerError, rw.Code)
		require.Equal(t, serverError, rw.Body.String())
	})

	t.Run("Base path not found in index", func(t *testing.T) {
		fileIndexDoc := &FileIndexDoc{
			ID: "file:idx:1234",
		}

		doc, err := getDocument(fileIndexDoc)
		require.NoError(t, err)

		docResolver.ResolveDocumentReturns(&document.ResolutionResult{Document: doc}, nil)

		getResourceName = func(req *http.Request) string { return schema1 }
		rw := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/schema/schema1.json", nil)
		h.Handler()(rw, req)
		require.Equal(t, http.StatusInternalServerError, rw.Code)
		require.Equal(t, serverError, rw.Body.String())
	})

	t.Run("File not found in index", func(t *testing.T) {
		fileIndexDoc := &FileIndexDoc{
			ID: "file:idx:1234",
			FileIndex: FileIndex{
				BasePath: "/schema",
				Mappings: nil,
			},
		}

		doc, err := getDocument(fileIndexDoc)
		require.NoError(t, err)

		docResolver.ResolveDocumentReturns(&document.ResolutionResult{Document: doc}, nil)

		getResourceName = func(req *http.Request) string { return schema1 }
		rw := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/schema/schema1.json", nil)
		h.Handler()(rw, req)
		require.Equal(t, http.StatusNotFound, rw.Code)
		require.Equal(t, fileNotFound, rw.Body.String())
	})

	t.Run("File not found in DCAS", func(t *testing.T) {
		fileIndexDoc := &FileIndexDoc{
			ID: "file:idx:1234",
			FileIndex: FileIndex{
				BasePath: "/schema",
				Mappings: map[string]string{
					"schema1.json": "1234567890",
				},
			},
		}

		doc, err := getDocument(fileIndexDoc)
		require.NoError(t, err)

		docResolver.ResolveDocumentReturns(&document.ResolutionResult{Document: doc}, nil)

		getResourceName = func(req *http.Request) string { return schema1 }
		rw := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/schema/schema1.json", nil)
		h.Handler()(rw, req)
		require.Equal(t, http.StatusNotFound, rw.Code)
		require.Equal(t, fileNotFound, rw.Body.String())
	})

	t.Run("DCAS provider error", func(t *testing.T) {
		dcasProvider.GetDCASClientReturns(nil, errors.New("injected DCAS provider error"))
		defer func() { dcasProvider.GetDCASClientReturns(dcasClient, nil) }()

		fileIndexDoc := &FileIndexDoc{
			ID: "file:idx:1234",
			FileIndex: FileIndex{
				BasePath: "/schema",
				Mappings: map[string]string{
					"schema1.json": "1234567890",
				},
			},
		}

		doc, err := getDocument(fileIndexDoc)
		require.NoError(t, err)

		docResolver.ResolveDocumentReturns(&document.ResolutionResult{Document: doc}, nil)
		dcasClient.WithGetError(errors.New("injected DCAS error"))
		defer dcasClient.WithGetError(nil)

		getResourceName = func(req *http.Request) string { return schema1 }
		rw := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/schema/schema1.json", nil)
		h.Handler()(rw, req)
		require.Equal(t, http.StatusInternalServerError, rw.Code)
		require.Equal(t, serverError, rw.Body.String())
	})

	t.Run("DCAS error", func(t *testing.T) {
		fileIndexDoc := &FileIndexDoc{
			ID: "file:idx:1234",
			FileIndex: FileIndex{
				BasePath: "/schema",
				Mappings: map[string]string{
					"schema1.json": "1234567890",
				},
			},
		}

		doc, err := getDocument(fileIndexDoc)
		require.NoError(t, err)

		docResolver.ResolveDocumentReturns(&document.ResolutionResult{Document: doc}, nil)
		dcasClient.WithGetError(errors.New("injected DCAS error"))
		defer dcasClient.WithGetError(nil)

		getResourceName = func(req *http.Request) string { return schema1 }
		rw := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/schema/schema1.json", nil)
		h.Handler()(rw, req)
		require.Equal(t, http.StatusInternalServerError, rw.Code)
		require.Equal(t, serverError, rw.Body.String())
	})

	t.Run("File retrieved from DCAS", func(t *testing.T) {
		fileIndexDoc := &FileIndexDoc{
			ID: "file:idx:1234",
			FileIndex: FileIndex{
				BasePath: "/schema",
				Mappings: map[string]string{
					"schema1.json": "1234567890",
				},
			},
		}

		doc, err := getDocument(fileIndexDoc)
		require.NoError(t, err)

		docResolver.ResolveDocumentReturns(&document.ResolutionResult{Document: doc}, nil)
		getResourceName = func(req *http.Request) string { return schema1 }

		t.Run("Invalid file", func(t *testing.T) {
			dcasClient.WithData("1234567890", []byte("{"))

			rw := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/schema/schema1.json", nil)
			h.Handler()(rw, req)
			require.Equal(t, http.StatusInternalServerError, rw.Code)
			require.Contains(t, rw.Body.String(), serverError)
		})

		t.Run("Missing content-type", func(t *testing.T) {
			dcasClient.WithData("1234567890", []byte("{}"))

			rw := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/schema/schema1.json", nil)
			h.Handler()(rw, req)
			require.Equal(t, http.StatusInternalServerError, rw.Code)
			require.Contains(t, rw.Body.String(), serverError)
		})

		t.Run("Empty content", func(t *testing.T) {
			f := &File{
				ContentType: "application/json",
			}

			fileBytes, err := json.Marshal(f)
			require.NoError(t, err)

			dcasClient.WithData("1234567890", fileBytes)

			rw := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/schema/schema1.json", nil)
			h.Handler()(rw, req)
			require.Equal(t, http.StatusNotFound, rw.Code)
		})

		t.Run("Success", func(t *testing.T) {
			fileContents := `{"field1":"value1"}`
			f := &File{
				ContentType: "application/json",
				Content:     []byte(fileContents),
			}

			fileBytes, err := json.Marshal(f)
			require.NoError(t, err)

			dcasClient.WithData("1234567890", fileBytes)

			rw := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/schema/schema1.json", nil)
			h.Handler()(rw, req)
			require.Equal(t, http.StatusOK, rw.Code)
			require.Equal(t, fileContents, rw.Body.String())
		})
	})
}

func getDocument(fileIndexDoc *FileIndexDoc) (document.Document, error) {
	bytes, err := json.Marshal(fileIndexDoc)
	if err != nil {
		return nil, err
	}

	doc := make(document.Document)
	err = json.Unmarshal(bytes, &doc)
	if err != nil {
		return nil, err
	}

	return doc, nil
}
