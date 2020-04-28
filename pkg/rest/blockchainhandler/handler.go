/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package blockchainhandler

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	cb "github.com/hyperledger/fabric-protos-go/common"
	bcclient "github.com/trustbloc/sidetree-fabric/pkg/client"
	"github.com/trustbloc/sidetree-fabric/pkg/httpserver"
)

type handler struct {
	Config
	path               string
	method             string
	params             map[string]string
	channelID          string
	blockchainProvider blockchainClientProvider
	jsonMarshal        func(v interface{}) ([]byte, error)
}

func newHandler(channelID string, cfg Config, path string, method string, blockchainProvider blockchainClientProvider, params ...string) *handler {
	return &handler{
		Config:             cfg,
		channelID:          channelID,
		path:               path,
		method:             method,
		params:             paramsBuilder(params).build(),
		blockchainProvider: blockchainProvider,
		jsonMarshal:        json.Marshal,
	}
}

// Path returns the context path
func (h *handler) Path() string {
	return h.path
}

// Params returns the accepted parameters
func (h *handler) Params() map[string]string {
	return h.params
}

// Method returns the HTTP method
func (h *handler) Method() string {
	return h.method
}

func (h *handler) blockchainClient() (bcclient.Blockchain, error) {
	bcClient, err := h.blockchainProvider.ForChannel(h.channelID)
	if err != nil {
		logger.Errorf("[%s] Failed to get blockchain client: %s", h.channelID, err)

		return nil, httpserver.ServerError
	}

	return bcClient, nil
}

func (h *handler) getBlockByHash(strHash string) (*cb.Block, error) {
	hash, err := base64.URLEncoding.DecodeString(strHash)
	if err != nil {
		logger.Debugf("Invalid base64 encoded hash [%s]: %s", strHash, err)

		return nil, httpserver.NewError(http.StatusBadRequest, httpserver.StatusBadRequest)
	}

	bcClient, err := h.blockchainClient()
	if err != nil {
		return nil, err
	}

	block, err := bcClient.GetBlockByHash(hash)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, httpserver.NotFoundError
		}

		logger.Errorf("Failed to get block for hash [%s]: %s", strHash, err)

		return nil, httpserver.ServerError
	}

	return block, nil
}

func (h *handler) getBlockchainInfo() (*cb.BlockchainInfo, error) {
	bcClient, err := h.blockchainClient()
	if err != nil {
		return nil, err
	}

	bcInfo, err := bcClient.GetBlockchainInfo()
	if err != nil {
		logger.Errorf("[%s] Failed to get blockchain info: %s", h.channelID, err)

		return nil, httpserver.ServerError
	}

	return bcInfo, nil
}

type paramsBuilder []string

func (p paramsBuilder) build() map[string]string {
	m := make(map[string]string)

	for _, p := range p {
		m[p] = fmt.Sprintf("{%s}", p)
	}

	return m
}
