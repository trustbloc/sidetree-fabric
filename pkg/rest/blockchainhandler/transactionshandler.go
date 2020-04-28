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
	"strings"

	"github.com/gorilla/mux"
	cb "github.com/hyperledger/fabric-protos-go/common"
	"github.com/pkg/errors"
	"github.com/trustbloc/sidetree-core-go/pkg/restapi/common"

	bcclient "github.com/trustbloc/sidetree-fabric/pkg/client"
	"github.com/trustbloc/sidetree-fabric/pkg/httpserver"
)

const (
	sinceParam    = "since"
	timeHashParam = "transaction_time_hash"
)

type getTransactionsFunc func(*http.Request) (*TransactionsResponse, error)

// Transactions retrieves the Sidetree transactions from the ledger
type Transactions struct {
	*handler
	getTransactions getTransactionsFunc
}

// NewTransactionsHandler returns a new blockchain Transactions handler
func NewTransactionsHandler(channelID string, cfg Config, blockchainProvider blockchainClientProvider) *Transactions {
	t := &Transactions{
		handler: newHandler(
			channelID, cfg,
			fmt.Sprintf("%s/transactions", cfg.BasePath),
			http.MethodGet,
			blockchainProvider,
		),
	}

	t.getTransactions = t.allTransactions

	return t
}

// NewTransactionsSinceHandler returns a new blockchain Transactions handler which returns all Sidetree transactions
// since the given block hash/transaction number
func NewTransactionsSinceHandler(channelID string, cfg Config, blockchainProvider blockchainClientProvider) *Transactions {
	t := &Transactions{
		handler: newHandler(
			channelID, cfg,
			fmt.Sprintf("%s/transactions", cfg.BasePath),
			http.MethodGet,
			blockchainProvider,
			sinceParam, timeHashParam,
		),
	}

	t.getTransactions = t.transactionsSince

	return t
}

// Handler returns the request handler
func (h *Transactions) Handler() common.HTTPRequestHandler {
	return h.transactions
}

func (h *Transactions) transactions(w http.ResponseWriter, req *http.Request) {
	rw := newBlockchainWriter(w)

	resp, err := h.getTransactions(req)
	if err != nil {
		rw.WriteError(err)
		return
	}

	transactionBytes, err := h.jsonMarshal(resp)
	if err != nil {
		logger.Errorf("[%s] Error marshalling transactions: %s", h.channelID, err)

		rw.WriteError(httpserver.ServerError)
		return
	}

	rw.Write(transactionBytes)
}

func (h *Transactions) allTransactions(req *http.Request) (*TransactionsResponse, error) {
	bcClient, err := h.blockchainClient()
	if err != nil {
		return nil, err
	}

	logger.Debugf("[%s] Returning all transactions since inception (max=%d) ...", h.channelID, h.MaxTransactionsInResponse)

	bcInfo, err := h.getBlockchainInfo()
	if err != nil {
		return nil, errors.WithMessage(err, "failed to get blockchain info")
	}

	resp, err := scanBlockchain(h.channelID, h.MaxTransactionsInResponse, newBlockScanner(h.channelID, bcClient), newBlockIterator(bcInfo, 1, 0))
	if err != nil {
		logger.Errorf("[%s] Failed to process blocks: %s", h.channelID, err)

		return nil, httpserver.ServerError
	}

	logger.Debugf("[%s] Returning: %+v", h.channelID, resp)

	return resp, nil
}

func (h *Transactions) transactionsSince(req *http.Request) (*TransactionsResponse, error) {
	strSince := getSince(req)
	sinceTxnNum, err := strconv.ParseUint(strSince, 10, 64)
	if err != nil {
		logger.Debugf("[%s] Invalid 'since' parameter [%s]: %s", h.channelID, strSince, err)

		return nil, newBadRequestError(InvalidTxNumOrTimeHash)
	}

	bcClient, err := h.blockchainClient()
	if err != nil {
		return nil, err
	}

	block, err := h.getBlockByHash(getTimeHash(req), bcClient)
	if err != nil {
		return nil, err
	}

	logger.Debugf("[%s] Returning transactions since block %d and TxNum %d (max=%d)...", h.channelID, block.Header.Number, sinceTxnNum, h.MaxTransactionsInResponse)

	bcInfo, err := bcClient.GetBlockchainInfo()
	if err != nil {
		return nil, errors.WithMessage(err, "failed to get blockchain info")
	}

	resp, err := scanBlockchain(h.channelID, h.MaxTransactionsInResponse, newBlockScanner(h.channelID, bcClient), newBlockIterator(bcInfo, block.Header.Number, sinceTxnNum))
	if err != nil {
		logger.Errorf("[%s] Failed to Scan blocks: %s", h.channelID, err)

		return nil, httpserver.ServerError
	}

	logger.Debugf("[%s] Returning: %+v", h.channelID, resp)

	return resp, nil
}

func (h *Transactions) blockchainClient() (bcclient.Blockchain, error) {
	bcClient, err := h.blockchainProvider.ForChannel(h.channelID)
	if err != nil {
		logger.Errorf("[%s] Failed to get blockchain client: %s", h.channelID, err)

		return nil, httpserver.ServerError
	}

	return bcClient, nil
}

func (h *Transactions) getBlockByHash(strHash string, bcClient bcclient.Blockchain) (*cb.Block, error) {
	hash, err := base64.URLEncoding.DecodeString(strHash)
	if err != nil {
		logger.Debugf("Invalid base64 encoded hash [%s]: %s", strHash, err)

		return nil, newBadRequestError(InvalidTxNumOrTimeHash)
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

var getSince = func(req *http.Request) string {
	return mux.Vars(req)[sinceParam]
}

var getTimeHash = func(req *http.Request) string {
	return mux.Vars(req)[timeHashParam]
}
