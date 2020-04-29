/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package blockchainhandler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/trustbloc/sidetree-core-go/pkg/restapi/common"

	"github.com/trustbloc/sidetree-fabric/pkg/httpserver"
)

type getBlocksFunc func(req *http.Request) ([]Block, error)

// Blocks retrieves blocks from the ledger
type Blocks struct {
	*handler
	getBlocks getBlocksFunc
}

// NewBlocksFromNumHandler returns a new handler that retrieves a range of blocks starting
// at the block number given by the parameter, from-time, and returns a maximum
// number of blocks given by parameter max-blocks. The blocks are returned in JSON encoding.
func NewBlocksFromNumHandler(channelID string, cfg Config, blockchainProvider blockchainClientProvider) *Blocks {
	h := &Blocks{
		handler: newHandler(
			channelID, cfg,
			fmt.Sprintf("%s/blocks", cfg.BasePath),
			http.MethodGet,
			blockchainProvider,
			fromTimeParam, maxBlocksParam,
		),
	}

	h.getBlocks = h.getBlocksFromNumber

	return h
}

// NewBlocksFromNumHandlerWithEncoding returns a new handler that retrieves a range of blocks starting
// at the block number given by the parameter, from-time, and returns a maximum
// number of blocks given by parameter max-blocks. The blocks are returned with data encoded in the provided encoding.
func NewBlocksFromNumHandlerWithEncoding(channelID string, cfg Config, blockchainProvider blockchainClientProvider) *Blocks {
	h := &Blocks{
		handler: newHandler(
			channelID, cfg,
			fmt.Sprintf("%s/blocks", cfg.BasePath),
			http.MethodGet,
			blockchainProvider,
			fromTimeParam, maxBlocksParam, dataEncodingParam,
		),
	}

	h.getBlocks = h.getBlocksFromNumber

	return h
}

// NewBlockByHashHandler returns a handler that retrieves a block by hash
func NewBlockByHashHandler(channelID string, cfg Config, blockchainProvider blockchainClientProvider) *Blocks {
	h := &Blocks{
		handler: newHandler(
			channelID, cfg,
			fmt.Sprintf("%s/blocks/{%s}", cfg.BasePath, hashParam),
			http.MethodGet,
			blockchainProvider,
		),
	}

	h.getBlocks = h.getBlockForHash

	return h
}

// NewBlockByHashHandlerWithEncoding returns a handler that retrieves a block by hash and returns the block
// with data encoded in the provided encoding.
func NewBlockByHashHandlerWithEncoding(channelID string, cfg Config, blockchainProvider blockchainClientProvider) *Blocks {
	h := &Blocks{
		handler: newHandler(
			channelID, cfg,
			fmt.Sprintf("%s/blocks/{%s}", cfg.BasePath, hashParam),
			http.MethodGet,
			blockchainProvider,
			dataEncodingParam,
		),
	}

	h.getBlocks = h.getBlockForHash

	return h
}

// Handler returns the request handler
func (h *Blocks) Handler() common.HTTPRequestHandler {
	return h.blocks
}

func (h *Blocks) blocks(w http.ResponseWriter, req *http.Request) {
	rw := newBlockchainWriter(w)

	blocks, err := h.getBlocks(req)
	if err != nil {
		rw.WriteError(err)
		return
	}

	blocksBytes, err := h.jsonMarshal(blocks)
	if err != nil {
		logger.Errorf("Unable to marshal blockchain blocks: %s", err)

		rw.WriteError(httpserver.ServerError)
		return
	}

	logger.Debugf("[%s] ... returning blockchain blocks: %s", h.channelID, blocksBytes)

	rw.Write(blocksBytes)
}

func (h *Blocks) getBlocksFromNumber(req *http.Request) ([]Block, error) {
	fromBlockNum, maxBlocks, encoding, err := h.getFromParams(req)
	if err != nil {
		return nil, err
	}

	bcInfo, err := h.getBlockchainInfo()
	if err != nil {
		return nil, err
	}

	if fromBlockNum >= bcInfo.Height {
		logger.Debugf("[%s] FromBlockNum [%d] is greater than the last block number [%d]", h.channelID, fromBlockNum, bcInfo.Height-1)

		return nil, httpserver.NotFoundError
	}

	toBlockNum := min(bcInfo.Height-1, fromBlockNum+uint64(maxBlocks)-1)

	blocks, err := h.getBlocksInRange(fromBlockNum, toBlockNum, encoding)
	if err != nil {
		logger.Errorf("[%s] Error retrieving blocks: %s", h.channelID, err)

		return nil, httpserver.ServerError
	}

	return blocks, nil
}

func (h *Blocks) getBlockForHash(req *http.Request) ([]Block, error) {
	hash, encoding, err := h.getByHashParams(req)
	if err != nil {
		return nil, err
	}

	block, err := h.getBlockByHash(hash)
	if err != nil {
		return nil, err
	}

	logger.Debugf("[%s] Got block for hash [%s]: %s", h.channelID, hash, block.Header)

	b, err := newBlockEncoder(encoding).encode(block)
	if err != nil {
		logger.Errorf("[%s] Error marshalling block to JSON: %s", h.channelID, err)

		return nil, httpserver.ServerError
	}

	return []Block{b}, nil
}

func (h *Blocks) getBlocksInRange(fromBlockNum, toBlockNum uint64, encoding DataEncoding) ([]Block, error) {
	logger.Debugf("[%s] Getting blocks from [%d] to [%d]", h.channelID, fromBlockNum, toBlockNum)

	blocks := make([]Block, toBlockNum-fromBlockNum+1)

	i := 0

	encoder := newBlockEncoder(encoding)

	for num := fromBlockNum; num <= toBlockNum; num++ {
		block, err := h.getBlockByNum(num)
		if err != nil {
			return nil, err
		}

		b, err := encoder.encode(block)
		if err != nil {
			return nil, err
		}

		blocks[i] = b
		i++
	}

	return blocks, nil
}

func (h *Blocks) getByHashParams(req *http.Request) (hash string, encoding DataEncoding, err error) {
	hash = getHash(req)
	if hash == "" {
		return "", "", newBadRequestError(InvalidTimeHash)
	}

	encoding = h.dataEncoding(req)

	return
}

func (h *Blocks) getFromParams(req *http.Request) (fromBlockNum uint64, maxBlocks int, encoding DataEncoding, err error) {
	strFromBlockNum := getFrom(req)
	if strFromBlockNum == "" {
		return 0, 0, "", newBadRequestError(InvalidFromTime)
	}

	fromBlockNum, err = strconv.ParseUint(strFromBlockNum, 10, 64)
	if err != nil {
		return 0, 0, "", newBadRequestError(InvalidFromTime)
	}

	strMaxBlocks := getMaxBlocks(req)
	if strMaxBlocks == "" {
		return 0, 0, "", newBadRequestError(InvalidMaxBlocks)
	}

	maxBlocks, err = strconv.Atoi(strMaxBlocks)
	if err != nil {
		return 0, 0, "", newBadRequestError(InvalidMaxBlocks)
	}

	if maxBlocks > h.MaxBlocksInResponse {
		logger.Debugf("[%s] Param, maxBlocks [%d] is greater than the allowable maxBlocks [%d]. Setting to [%d]", h.channelID, maxBlocks, h.MaxBlocksInResponse, h.MaxBlocksInResponse)
		maxBlocks = h.MaxBlocksInResponse
	}

	encoding = h.dataEncoding(req)

	logger.Debugf("[%s] Using params - fromBlockNum=%d, maxBlocks=%d, encoding=[%s]", h.channelID, fromBlockNum, maxBlocks, encoding)

	return
}

var getFrom = func(req *http.Request) string {
	return mux.Vars(req)[fromTimeParam]
}

var getMaxBlocks = func(req *http.Request) string {
	return mux.Vars(req)[maxBlocksParam]
}

func min(v1, v2 uint64) uint64 {
	if v1 < v2 {
		return v1
	}

	return v2
}
