/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package sidetreesvc

import (
	"github.com/trustbloc/sidetree-core-go/pkg/restapi/common"
)

type fileHandlers struct {
	readHandler  common.HTTPHandler
	writeHandler common.HTTPHandler
}

// HTTPHandlers returns the HTTP handlers
func (h *fileHandlers) HTTPHandlers() []common.HTTPHandler {
	var handlers []common.HTTPHandler

	if h.readHandler != nil {
		handlers = append(handlers, h.readHandler)
	}

	if h.writeHandler != nil {
		handlers = append(handlers, h.writeHandler)
	}

	return handlers
}

type dcasHandlers struct {
	readHandler    common.HTTPHandler
	writeHandler   common.HTTPHandler
	versionHandler common.HTTPHandler
}

// HTTPHandlers returns the HTTP handlers
func (h *dcasHandlers) HTTPHandlers() []common.HTTPHandler {
	var handlers []common.HTTPHandler

	if h.readHandler != nil {
		handlers = append(handlers, h.readHandler)
	}

	if h.writeHandler != nil {
		handlers = append(handlers, h.writeHandler)
	}

	if h.versionHandler != nil {
		handlers = append(handlers, h.versionHandler)
	}

	return handlers
}

type blockchainHandlers struct {
	timeHandler       common.HTTPHandler
	timeByHashHandler common.HTTPHandler
}

func (h *blockchainHandlers) HTTPHandlers() []common.HTTPHandler {
	var handlers []common.HTTPHandler

	if h.timeHandler != nil {
		handlers = append(handlers, h.timeHandler)
	}

	if h.timeByHashHandler != nil {
		handlers = append(handlers, h.timeByHashHandler)
	}

	return handlers
}
