/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package blockchainhandler

import (
	"net/http"

	"github.com/trustbloc/sidetree-fabric/pkg/httpserver"
)

type responseWriter struct {
	*httpserver.ResponseWriter
}

func newBlockchainWriter(rw http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: httpserver.NewResponseWriter(rw),
	}
}

func (rw *responseWriter) Write(content []byte) {
	rw.ResponseWriter.Write(http.StatusOK, content, httpserver.ContentTypeJSON)
}
