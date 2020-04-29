/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package blockchainhandler

import (
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/trustbloc/sidetree-core-go/pkg/restapi/common"
	"github.com/trustbloc/sidetree-fabric/pkg/httpserver"
)

// Info retrieves basic information about the blockchain
type Info struct {
	*handler
}

// NewInfoHandler returns a new blockchain info handler
func NewInfoHandler(channelID string, cfg Config, blockchainProvider blockchainClientProvider) *Info {
	return &Info{
		handler: newHandler(
			channelID, cfg,
			fmt.Sprintf("%s/info", cfg.BasePath),
			http.MethodGet,
			blockchainProvider,
		),
	}
}

// Handler returns the request handler
func (h *Info) Handler() common.HTTPRequestHandler {
	return h.blockchainInfo
}

func (h *Info) blockchainInfo(w http.ResponseWriter, _ *http.Request) {
	rw := httpserver.NewResponseWriter(w)

	bcInfo, err := h.getBlockchainInfo()
	if err != nil {
		rw.WriteError(err)
		return
	}

	resp := &InfoResponse{
		CurrentTime:      bcInfo.Height - 1,
		CurrentTimeHash:  base64.URLEncoding.EncodeToString(bcInfo.CurrentBlockHash),
		PreviousTimeHash: base64.URLEncoding.EncodeToString(bcInfo.PreviousBlockHash),
	}

	infoBytes, err := h.jsonMarshal(resp)
	if err != nil {
		logger.Errorf("Unable to marshal blockchain info: %s", err)

		rw.WriteError(httpserver.ServerError)
		return
	}

	logger.Debugf("[%s] ... returning blockchain info: %s", h.channelID, infoBytes)

	rw.Write(http.StatusOK, infoBytes, httpserver.ContentTypeJSON)
}
