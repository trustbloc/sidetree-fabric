/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package blockchainhandler

import (
	"fmt"
	"net/http"

	cb "github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric/protoutil"
	"github.com/trustbloc/sidetree-core-go/pkg/restapi/common"
	"github.com/trustbloc/sidetree-fabric/pkg/httpserver"
)

type getConfigBlockFunc func(req *http.Request) (Block, error)

// ConfigBlock retrieves basic information about the Fabric blockchain
type ConfigBlock struct {
	*handler
	getConfigBlock getConfigBlockFunc
}

// NewConfigBlockHandler returns a new config block handler which returns the latest config block
func NewConfigBlockHandler(channelID string, cfg Config, blockchainProvider blockchainClientProvider) *ConfigBlock {
	h := newConfigBlock(
		channelID, cfg,
		fmt.Sprintf("%s/config-block", cfg.BasePath),
		blockchainProvider,
	)

	h.getConfigBlock = h.getLatestConfigBlock

	return h
}

// NewConfigBlockHandlerWithEncoding returns a new config block handler which returns the latest config block
func NewConfigBlockHandlerWithEncoding(channelID string, cfg Config, blockchainProvider blockchainClientProvider) *ConfigBlock {
	h := newConfigBlock(
		channelID, cfg,
		fmt.Sprintf("%s/config-block", cfg.BasePath),
		blockchainProvider,
		dataEncodingParam,
	)

	h.getConfigBlock = h.getLatestConfigBlock

	return h
}

// NewConfigBlockByHashHandler returns a new config block handler which returns the config block that
// was used by the block with the provided block hash
func NewConfigBlockByHashHandler(channelID string, cfg Config, blockchainProvider blockchainClientProvider) *ConfigBlock {
	h := newConfigBlock(
		channelID, cfg,
		fmt.Sprintf("%s/config-block/{%s}", cfg.BasePath, hashParam),
		blockchainProvider,
	)

	h.getConfigBlock = h.getConfigBlockByHash

	return h
}

// NewConfigBlockByHashHandlerWithEncoding returns a new config block handler which returns the config block that
// was used by the block with the provided block hash
func NewConfigBlockByHashHandlerWithEncoding(channelID string, cfg Config, blockchainProvider blockchainClientProvider) *ConfigBlock {
	h := newConfigBlock(
		channelID, cfg,
		fmt.Sprintf("%s/config-block/{%s}", cfg.BasePath, hashParam),
		blockchainProvider,
		dataEncodingParam,
	)

	h.getConfigBlock = h.getConfigBlockByHash

	return h
}

func newConfigBlock(channelID string, cfg Config, path string, blockchainProvider blockchainClientProvider, params ...string) *ConfigBlock {
	return &ConfigBlock{
		handler: newHandler(
			channelID, cfg, path, http.MethodGet,
			blockchainProvider, params...,
		),
	}
}

// Handler returns the request handler
func (h *ConfigBlock) Handler() common.HTTPRequestHandler {
	return h.configBlock
}

func (h *ConfigBlock) configBlock(w http.ResponseWriter, req *http.Request) {
	rw := newBlockchainWriter(w)

	block, err := h.getConfigBlock(req)
	if err != nil {
		rw.WriteError(err)
		return
	}

	infoBytes, err := h.jsonMarshal(block)
	if err != nil {
		logger.Errorf("Unable to marshal config block: %s", err)

		rw.WriteError(httpserver.ServerError)
		return
	}

	logger.Debugf("[%s] ... returning config block: %s", h.channelID, infoBytes)

	rw.Write(infoBytes)
}

func (h *ConfigBlock) getLatestConfigBlock(req *http.Request) (Block, error) {
	bcInfo, err := h.getBlockchainInfo()
	if err != nil {
		return nil, err
	}

	block, err := h.getBlockByNum(bcInfo.Height - 1)
	if err != nil {
		logger.Errorf("[%s] Failed to get latest block %d: %s", h.channelID, bcInfo.Height-1, err)

		return nil, httpserver.ServerError
	}

	return h.getConfigBlockFromMetadata(block, h.dataEncoding(req))
}

func (h *ConfigBlock) getConfigBlockByHash(req *http.Request) (Block, error) {
	hash, encoding, err := h.getBlockByHashParams(req)
	if err != nil {
		return nil, err
	}

	block, err := h.getBlockByHash(hash)
	if err != nil {
		return nil, err
	}

	return h.getConfigBlockFromMetadata(block, encoding)
}

func (h *ConfigBlock) getConfigBlockFromMetadata(block *cb.Block, encoding DataEncoding) (Block, error) {
	configBlockNum, err := protoutil.GetLastConfigIndexFromBlock(block)
	if err != nil {
		logger.Errorf("[%s] Error getting config index from block: %s", h.channelID, err)

		return nil, httpserver.ServerError
	}

	logger.Debugf("[%s] Got config block number [%d]", h.channelID, configBlockNum)

	configBlock, err := h.getBlockByNum(configBlockNum)
	if err != nil {
		return nil, err
	}

	tree, err := newBlockEncoder(encoding).encode(configBlock)
	if err != nil {
		return nil, err
	}

	return tree, nil
}

func (h *ConfigBlock) getBlockByHashParams(req *http.Request) (hash string, encoding DataEncoding, err error) {
	hash = getHash(req)
	if hash == "" {
		return "", "", newBadRequestError(InvalidTimeHash)
	}

	encoding = h.dataEncoding(req)

	logger.Debugf("[%s] Using params - hash=[%s], encoding=[%s]", h.channelID, hash, encoding)

	return
}
