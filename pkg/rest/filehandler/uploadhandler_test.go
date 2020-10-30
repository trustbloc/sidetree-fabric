/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package filehandler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"

	"github.com/trustbloc/sidetree-fabric/pkg/mocks"
)

func TestFileUploadHandler(t *testing.T) {
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

	h := NewUploadHandler(channelID, cfg, dcasProvider)
	require.Equal(t, cfg.BasePath, h.Path())
	require.Equal(t, http.MethodPost, h.Method())
	require.NotNil(t, h.Handler())

	t.Run("Bad request", func(t *testing.T) {
		rw := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/schema", nil)
		h.Handler()(rw, req)
		require.Equal(t, http.StatusBadRequest, rw.Code)
		require.Equal(t, "bad request", rw.Body.String())
	})

	t.Run("No content type", func(t *testing.T) {
		fileBytes, err := json.Marshal(&File{})
		require.NoError(t, err)

		rw := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/schema", bytes.NewReader(fileBytes))
		h.Handler()(rw, req)
		require.Equal(t, http.StatusBadRequest, rw.Code)
		require.Equal(t, "content type is required", rw.Body.String())
	})

	t.Run("No content type", func(t *testing.T) {
		fileBytes, err := json.Marshal(&File{
			ContentType: "application/json",
		})
		require.NoError(t, err)

		rw := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/schema", bytes.NewReader(fileBytes))
		h.Handler()(rw, req)
		require.Equal(t, http.StatusBadRequest, rw.Code)
		require.Equal(t, "content is required", rw.Body.String())
	})

	t.Run("DCAS provider error", func(t *testing.T) {
		dcasProvider.GetDCASClientReturns(nil, errors.New("injected DCAS provider error"))
		defer func() { dcasProvider.GetDCASClientReturns(dcasClient, nil) }()

		content := `{"field1","value1"}`
		f := &File{
			ContentType: "application/json",
			Content:     []byte(content),
		}

		fileBytes, err := json.Marshal(f)
		require.NoError(t, err)

		rw := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/schema", bytes.NewReader(fileBytes))
		h.Handler()(rw, req)
		require.Equal(t, http.StatusInternalServerError, rw.Code)
	})

	t.Run("DCAS error", func(t *testing.T) {
		dcasClient.WithPutError(errors.New("injected DCAS error"))
		defer dcasClient.WithPutError(nil)

		content := `{"field1","value1"}`
		f := &File{
			ContentType: "application/json",
			Content:     []byte(content),
		}

		fileBytes, err := json.Marshal(f)
		require.NoError(t, err)

		rw := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/schema", bytes.NewReader(fileBytes))
		h.Handler()(rw, req)
		require.Equal(t, http.StatusInternalServerError, rw.Code)
	})

	t.Run("Success", func(t *testing.T) {
		content := `{"field1","value1"}`
		f := &File{
			ContentType: "application/json",
			Content:     []byte(content),
		}

		fileBytes, err := json.Marshal(f)
		require.NoError(t, err)

		rw := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/schema", bytes.NewReader(fileBytes))
		h.Handler()(rw, req)
		require.Equal(t, http.StatusOK, rw.Code)
	})
}
