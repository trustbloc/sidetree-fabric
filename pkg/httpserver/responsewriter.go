/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package httpserver

import (
	"net/http"
)

const (
	// ContentTypeHeader is the name of the content-type header field
	ContentTypeHeader = "Content-Type"
	// ContentTypeJSON is the JSON content-type
	ContentTypeJSON = "application/json"
	// ContentTypeText is the text content-type
	ContentTypeText = "text/plain"
	// ContentTypeBinary is the binary content type
	ContentTypeBinary = "application/octet-stream"
)

// ResponseWriter wraps the http response writer and implements utility functions
type ResponseWriter struct {
	http.ResponseWriter
}

// NewResponseWriter returns a new response writer
func NewResponseWriter(rw http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{
		ResponseWriter: rw,
	}
}

// Write writes the response
func (rw *ResponseWriter) Write(status int, bytes []byte, contentType string) {
	rw.Header().Set(ContentTypeHeader, contentType)
	rw.WriteHeader(status)

	logger.Debugf("Writing response - Status: %d, Payload: %s", status, bytes)

	if _, err := rw.ResponseWriter.Write(bytes); err != nil {
		logger.Errorf("Unable to write UploadResponse: %s", err)
	}
}

// WriteText writes the given text to the response writer
func (rw *ResponseWriter) WriteText(status int, text string) {
	rw.Header().Set(ContentTypeHeader, ContentTypeText)
	rw.WriteHeader(status)

	logger.Debugf("Writing text response - Status: %d, Payload: %s", status, text)

	if _, err := rw.ResponseWriter.Write([]byte(text)); err != nil {
		logger.Errorf("Unable to write response: %s", err)
	}
}

// WriteError writes the given error to the response writer
func (rw *ResponseWriter) WriteError(err error) {
	httpErr, ok := err.(*Error)
	if ok {
		rw.WriteText(httpErr.Status(), httpErr.StatusMsg())
		return
	}

	rw.WriteText(http.StatusInternalServerError, StatusServerError)
}
