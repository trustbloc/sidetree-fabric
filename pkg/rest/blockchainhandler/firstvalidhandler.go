/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package blockchainhandler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/trustbloc/sidetree-core-go/pkg/restapi/common"

	"github.com/trustbloc/sidetree-fabric/pkg/httpserver"
)

// FirstValid retrieves the first valid transaction in a given set of transactions
type FirstValid struct {
	*handler
}

// NewFirstValidHandler returns a new, first valid handler
func NewFirstValidHandler(channelID string, cfg Config, blockchainProvider blockchainClientProvider) *FirstValid {
	return &FirstValid{
		handler: newHandler(
			channelID, cfg,
			fmt.Sprintf("%s/first-valid", cfg.BasePath),
			http.MethodPost,
			blockchainProvider,
		),
	}
}

// Handler returns the request handler
func (h *FirstValid) Handler() common.HTTPRequestHandler {
	return h.firstValid
}

func (h *FirstValid) firstValid(w http.ResponseWriter, req *http.Request) {
	rw := newBlockchainWriter(w)

	transactions, err := h.transactions(req)
	if err != nil {
		rw.WriteError(err)
		return
	}

	resp, err := h.getFirstValid(transactions)
	if err != nil {
		rw.WriteError(err)
		return
	}

	transactionBytes, err := h.jsonMarshal(resp)
	if err != nil {
		logger.Errorf("[%s] Error marshalling response: %s", h.channelID, err)

		rw.WriteError(httpserver.ServerError)
		return
	}

	rw.Write(transactionBytes)
}

func (h *FirstValid) transactions(req *http.Request) ([]Transaction, error) {
	var transactions []Transaction
	err := json.NewDecoder(req.Body).Decode(&transactions)
	if err != nil {
		logger.Debugf("[%s] Failed to unmarshal transactions from request: %s", h.channelID, err)

		return nil, httpserver.BadRequestError
	}

	return transactions, nil
}

func (h *FirstValid) getFirstValid(transactions []Transaction) (*Transaction, error) {
	bcClient, err := h.blockchainClient()
	if err != nil {
		return nil, err
	}

	logger.Debugf("[%s] Returning the first valid transaction from the given set of transactions: %s", h.channelID, transactions)

	resp, err := scanBlockchain(
		h.channelID, 1,
		newFirstValidScanner(h.channelID, bcClient),
		newFirstValidIterator(transactions),
	)
	if err != nil {
		logger.Errorf("[%s] Failed to Scan blocks: %s", h.channelID, err)

		return nil, httpserver.ServerError
	}

	if len(resp.Transactions) == 0 {
		logger.Debugf("[%s] No valid transactions found", h.channelID)

		return nil, httpserver.NotFoundError
	}

	logger.Debugf("[%s] Returning: %+v", h.channelID, resp)

	return &resp.Transactions[0], nil
}
