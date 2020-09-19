/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package filehandler

import (
	"net/http"

	"github.com/trustbloc/sidetree-core-go/pkg/api/protocol"
	"github.com/trustbloc/sidetree-core-go/pkg/restapi/common"
	resthandler "github.com/trustbloc/sidetree-core-go/pkg/restapi/dochandler"
)

// UpdateIndex is a REST handler that creates/updates file index Sidetree documents
type UpdateIndex struct {
	*resthandler.UpdateHandler
	path string
}

// NewUpdateIndexHandler returns a new update index handler
func NewUpdateIndexHandler(path string, processor resthandler.Processor, pc protocol.Client) *UpdateIndex {
	return &UpdateIndex{
		path:          path,
		UpdateHandler: resthandler.NewUpdateHandler(processor, pc),
	}
}

// Path returns the context path
func (h *UpdateIndex) Path() string {
	return h.path
}

// Method returns the HTTP method
func (h *UpdateIndex) Method() string {
	return http.MethodPost
}

// Handler returns the handler
func (h *UpdateIndex) Handler() common.HTTPRequestHandler {
	return h.Update
}
