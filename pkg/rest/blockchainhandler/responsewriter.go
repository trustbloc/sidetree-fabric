/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package blockchainhandler

import (
	"encoding/json"
	"net/http"

	"github.com/trustbloc/sidetree-fabric/pkg/httpserver"
)

type responseWriter struct {
	*httpserver.ResponseWriter
	jsonMarshal func(v interface{}) ([]byte, error)
}

func newBlockchainWriter(rw http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: httpserver.NewResponseWriter(rw),
		jsonMarshal:    json.Marshal,
	}
}

func (rw *responseWriter) Write(content []byte) {
	rw.ResponseWriter.Write(http.StatusOK, content, httpserver.ContentTypeJSON)
}

func (rw *responseWriter) WriteError(err error) {
	bcErr, ok := err.(*blockchainError)
	if ok {
		rw.WriteBlockchainError(bcErr)
		return
	}

	rw.ResponseWriter.WriteError(err)
}

func (rw *responseWriter) WriteBlockchainError(err *blockchainError) {
	errResponse := &ErrorResponse{
		Code: err.ResultCode(),
	}

	responseBytes, e := rw.jsonMarshal(errResponse)
	if e != nil {
		logger.Errorf("Failed to marshal error response %+v: %s", errResponse, e)

		rw.ResponseWriter.Write(err.Status(), []byte(err.ResultCode()), httpserver.ContentTypeText)
		return
	}

	rw.ResponseWriter.Write(err.Status(), responseBytes, httpserver.ContentTypeJSON)
}
