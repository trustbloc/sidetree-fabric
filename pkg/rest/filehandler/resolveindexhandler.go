/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package filehandler

import (
	"net/http"

	"github.com/trustbloc/sidetree-core-go/pkg/restapi/common"
	"github.com/trustbloc/sidetree-core-go/pkg/restapi/dochandler"
)

// ResolveIndex is a REST handler that retrieves a file index Sidetree documents by ID
type ResolveIndex struct {
	*dochandler.ResolveHandler
	path string
}

// NewResolveIndexHandler returns a new resolve index handler
func NewResolveIndexHandler(path string, resolver dochandler.Resolver) *ResolveIndex {
	return &ResolveIndex{
		ResolveHandler: dochandler.NewResolveHandler(resolver),
		path:           path,
	}
}

// Path returns the context path
func (h *ResolveIndex) Path() string {
	return h.path + "/identifiers/{id}"
}

// Method returns the HTTP method
func (h *ResolveIndex) Method() string {
	return http.MethodGet
}

// Handler returns the handler
func (h *ResolveIndex) Handler() common.HTTPRequestHandler {
	return h.Resolve
}
