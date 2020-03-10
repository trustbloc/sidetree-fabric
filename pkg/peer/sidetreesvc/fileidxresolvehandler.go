/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package sidetreesvc

import (
	"net/http"

	"github.com/trustbloc/sidetree-core-go/pkg/restapi/common"
	"github.com/trustbloc/sidetree-core-go/pkg/restapi/dochandler"
)

type fileIdxResolveHandler struct {
	*dochandler.ResolveHandler
	path string
}

func newFileIdxResolveHandler(path string, resolver dochandler.Resolver) *fileIdxResolveHandler {
	return &fileIdxResolveHandler{
		ResolveHandler: dochandler.NewResolveHandler(resolver),
		path:           path,
	}
}

// Path returns the context path
func (h *fileIdxResolveHandler) Path() string {
	return h.path + "/{id}"
}

// Method returns the HTTP method
func (h *fileIdxResolveHandler) Method() string {
	return http.MethodGet
}

// Handler returns the handler
func (h *fileIdxResolveHandler) Handler() common.HTTPRequestHandler {
	return h.Resolve
}
