/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package blockchainhandler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/hyperledger/fabric/common/flogging"
	"github.com/pkg/errors"
	"github.com/trustbloc/sidetree-core-go/pkg/restapi/common"
	bcclient "github.com/trustbloc/sidetree-fabric/pkg/client"
	"github.com/trustbloc/sidetree-fabric/pkg/httpserver"
)

var logger = flogging.MustGetLogger("sidetree_peer")

// Time retrieves the blockchain time from the ledger (block height and latest block hash)
type Time struct {
	Config
	channelID          string
	blockchainProvider blockchainClientProvider
	jsonMarshal        func(v interface{}) ([]byte, error)
}

type blockchainClientProvider interface {
	ForChannel(channelID string) (bcclient.Blockchain, error)
}

// NewTimeHandler returns a new blockchain time handler
func NewTimeHandler(channelID string, cfg Config, blockchainProvider blockchainClientProvider) *Time {
	return &Time{
		Config:             cfg,
		channelID:          channelID,
		blockchainProvider: blockchainProvider,
		jsonMarshal:        json.Marshal,
	}
}

// Path returns the context path
func (h *Time) Path() string {
	return h.BasePath + "/time"
}

// Method returns the HTTP method
func (h *Time) Method() string {
	return http.MethodGet
}

// Handler returns the request handler
func (h *Time) Handler() common.HTTPRequestHandler {
	return h.time
}

func (h *Time) time(w http.ResponseWriter, req *http.Request) {
	rw := httpserver.NewResponseWriter(w)

	time, err := h.getTime()
	if err != nil {
		logger.Errorf("[%s] Unable to get blockchain time: %s", h.channelID, err)

		rw.WriteError(httpserver.ServerError)
		return
	}

	timeBytes, err := h.jsonMarshal(time)
	if err != nil {
		logger.Errorf("Unable to marshal blockchain time: %s", err)

		rw.WriteError(httpserver.ServerError)
		return
	}

	logger.Debugf("[%s] ... returning blockchain time: %s", h.channelID, timeBytes)

	rw.Write(http.StatusOK, timeBytes, httpserver.ContentTypeJSON)
}

func (h *Time) getTime() (*TimeResponse, error) {
	bcClient, err := h.blockchainProvider.ForChannel(h.channelID)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to get blockchain client")
	}

	bcInfo, err := bcClient.GetBlockchainInfo()
	if err != nil {
		return nil, errors.WithMessage(err, "failed to get blockchain info")
	}

	return &TimeResponse{
		Time: strconv.FormatUint(bcInfo.Height, 10),
		Hash: bcInfo.CurrentBlockHash,
	}, nil
}
