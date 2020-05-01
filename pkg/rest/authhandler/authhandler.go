/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package authhandler

import (
	"crypto/subtle"
	"net/http"

	"github.com/hyperledger/fabric/common/flogging"
	"github.com/trustbloc/sidetree-core-go/pkg/restapi/common"
)

var logger = flogging.MustGetLogger("sidetree_peer")

const (
	authHeader  = "Authorization"
	tokenPrefix = "Bearer "
)

// Handler authorizes the request before delegating to the target handler
type Handler struct {
	common.HTTPHandler
	channelID string
	tokens    []string
}

// New returns a new auth handler
func New(channelID string, tokens []string, handler common.HTTPHandler) common.HTTPHandler {
	if len(tokens) == 0 {
		logger.Debugf("[%s] No authorization token(s) specified. Authorization will NOT be performed for %s on path [%s]", channelID, handler.Method(), handler.Path())
		return handler
	}

	logger.Debugf("[%s] Creating authorization handler for %s on path [%s]", channelID, handler.Method(), handler.Path())

	return &Handler{
		HTTPHandler: handler,
		channelID:   channelID,
		tokens:      tokens,
	}
}

// Handler returns the HTTP request handler
func (h *Handler) Handler() common.HTTPRequestHandler {
	return h.handle
}

// Params returns the parameters
func (h *Handler) Params() map[string]string {
	ph, ok := h.HTTPHandler.(paramHolder)
	if ok {
		return ph.Params()
	}

	return nil
}

func (h *Handler) handle(w http.ResponseWriter, r *http.Request) {
	if !h.isAuthorized(r) {
		logger.Debugf("[%s] Caller is not authorized for %s on path [%s]", h.channelID, h.Method(), h.Path())

		w.WriteHeader(http.StatusUnauthorized)

		if _, err := w.Write([]byte("Unauthorized.\n")); err != nil {
			logger.Warnf("[%s] Failed to write response: %s", h.channelID, err)
		}

		return
	}

	logger.Debugf("[%s] Caller is authorized for %s on path [%s]", h.channelID, h.Method(), h.Path())

	h.HTTPHandler.Handler()(w, r)
}

func (h *Handler) isAuthorized(r *http.Request) bool {
	actHdr := r.Header.Get(authHeader)

	// Compare the header against all tokens. If any match then we allow the request.
	for _, token := range h.tokens {
		if subtle.ConstantTimeCompare([]byte(actHdr), []byte(tokenPrefix+token)) == 1 {
			return true
		}
	}

	return false
}

type paramHolder interface {
	Params() map[string]string
}
