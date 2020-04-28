/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package blockchainhandler

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/hyperledger/fabric/common/flogging"
	"github.com/hyperledger/fabric/protoutil"
	"github.com/trustbloc/sidetree-core-go/pkg/restapi/common"

	bcclient "github.com/trustbloc/sidetree-fabric/pkg/client"
	"github.com/trustbloc/sidetree-fabric/pkg/httpserver"
)

var logger = flogging.MustGetLogger("sidetree_peer")

const (
	hashParam = "hash"
)

type getTimeFunc func(req *http.Request) (*TimeResponse, error)

// Time retrieves the blockchain time from the ledger (block height and latest block hash)
type Time struct {
	*handler
	getTime getTimeFunc
}

type blockchainClientProvider interface {
	ForChannel(channelID string) (bcclient.Blockchain, error)
}

// NewTimeHandler returns a new blockchain time handler
func NewTimeHandler(channelID string, cfg Config, blockchainProvider blockchainClientProvider) *Time {
	t := &Time{
		handler: newHandler(
			channelID, cfg,
			fmt.Sprintf("%s/time", cfg.BasePath),
			http.MethodGet,
			blockchainProvider,
		),
	}

	t.getTime = t.getLatestTime

	return t
}

// NewTimeByHashHandler returns a new blockchain time handler for a given hash
func NewTimeByHashHandler(channelID string, cfg Config, blockchainProvider blockchainClientProvider) *Time {
	t := &Time{
		handler: newHandler(
			channelID, cfg,
			fmt.Sprintf("%s/time/{hash}", cfg.BasePath),
			http.MethodGet,
			blockchainProvider,
		),
	}

	t.getTime = t.getTimeForHash

	return t
}

// Handler returns the request handler
func (h *Time) Handler() common.HTTPRequestHandler {
	return h.time
}

func (h *Time) time(w http.ResponseWriter, req *http.Request) {
	rw := httpserver.NewResponseWriter(w)

	time, err := h.getTime(req)
	if err != nil {
		rw.WriteError(err)
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

func (h *Time) getLatestTime(req *http.Request) (*TimeResponse, error) {
	bcInfo, err := h.getBlockchainInfo()
	if err != nil {
		logger.Errorf("[%s] Failed to get blockchain info: %s", h.channelID, err)

		return nil, httpserver.ServerError
	}

	logger.Debugf("Got latest blockchain info: %s", bcInfo)

	return &TimeResponse{
		Time: strconv.FormatUint(bcInfo.Height-1, 10),
		Hash: base64.URLEncoding.EncodeToString(bcInfo.CurrentBlockHash),
	}, nil
}

func (h *Time) getTimeForHash(req *http.Request) (*TimeResponse, error) {
	strHash := getHash(req)
	if strHash == "" {
		return nil, httpserver.BadRequestError
	}

	block, err := h.getBlockByHash(strHash)
	if err != nil {
		return nil, err
	}

	header := block.Header
	if header == nil {
		logger.Errorf("[%s] Nil header in block", h.channelID)

		return nil, httpserver.ServerError
	}

	logger.Debugf("[%s] Got block header for hash [%s]: %s", h.channelID, strHash, header)

	return &TimeResponse{
		Time: strconv.FormatUint(header.Number, 10),
		Hash: base64.URLEncoding.EncodeToString(protoutil.BlockHeaderHash(header)),
	}, nil
}

var getHash = func(req *http.Request) string {
	return mux.Vars(req)[hashParam]
}
