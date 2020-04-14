/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package httpserver

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/trustbloc/sidetree-fabric/pkg/httpserver/mocks"
)

//go:generate counterfeiter -o ./mocks/responsewriter.gen.go --fake-name ResponseWriter net/http.ResponseWriter

func TestResponseWriter_Write(t *testing.T) {
	const content = `{"field":"value"}`

	r := httptest.NewRecorder()
	rw := NewResponseWriter(r)
	require.NotNil(t, rw)

	rw.Write(http.StatusOK, []byte(content), ContentTypeJSON)

	require.Equal(t, http.StatusOK, r.Result().StatusCode)
	require.Equal(t, content, r.Body.String())
	require.Equal(t, ContentTypeJSON, r.Header().Get(ContentTypeHeader))
}

func TestResponseWriter_WriteText(t *testing.T) {
	const content = "some text string"

	r := httptest.NewRecorder()
	rw := NewResponseWriter(r)
	require.NotNil(t, rw)

	rw.WriteText(http.StatusOK, content)

	require.Equal(t, http.StatusOK, r.Result().StatusCode)
	require.Equal(t, content, r.Body.String())
	require.Equal(t, ContentTypeText, r.Header().Get(ContentTypeHeader))
}

func TestResponseWriter_WriteError(t *testing.T) {
	errExpected := NewError(http.StatusBadRequest, StatusBadRequest)

	r := httptest.NewRecorder()
	rw := NewResponseWriter(r)
	require.NotNil(t, rw)

	rw.WriteError(errExpected)

	require.Equal(t, http.StatusBadRequest, r.Result().StatusCode)
	require.Equal(t, StatusBadRequest, r.Body.String())
	require.Equal(t, ContentTypeText, r.Header().Get(ContentTypeHeader))
}

func TestResponseWriter_WriterError(t *testing.T) {
	errExpected := errors.New("injected writer error")

	r := &mocks.ResponseWriter{}
	r.WriteReturns(0, errExpected)
	r.HeaderReturns(make(http.Header))

	rw := NewResponseWriter(r)
	require.NotNil(t, rw)

	require.NotPanics(t, func() { rw.Write(http.StatusOK, []byte(`{}`), ContentTypeJSON) })
	require.NotPanics(t, func() { rw.WriteError(errExpected) })
}
