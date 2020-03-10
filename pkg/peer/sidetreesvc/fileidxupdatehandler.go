/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package sidetreesvc

import (
	"net/http"

	"github.com/trustbloc/sidetree-core-go/pkg/restapi/common"
	resthandler "github.com/trustbloc/sidetree-core-go/pkg/restapi/dochandler"
)

type fileIdxUpdateHandler struct {
	*resthandler.UpdateHandler
	path string
}

func newFileIdxUpdateHandler(path string, processor resthandler.Processor) *fileIdxUpdateHandler {
	return &fileIdxUpdateHandler{
		path:          path,
		UpdateHandler: resthandler.NewUpdateHandler(processor),
	}
}

// Path returns the context path
func (h *fileIdxUpdateHandler) Path() string {
	return h.path
}

// Method returns the HTTP method
func (h *fileIdxUpdateHandler) Method() string {
	return http.MethodPost
}

// Handler returns the handler
func (h *fileIdxUpdateHandler) Handler() common.HTTPRequestHandler {
	return h.Update
}
