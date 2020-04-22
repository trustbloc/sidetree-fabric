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

	bcclient "github.com/trustbloc/sidetree-fabric/pkg/client"
	"github.com/trustbloc/sidetree-fabric/pkg/httpserver"
)

// FirstValid retrieves the first valid transaction in a given set of transactions
type FirstValid struct {
	Config
	path               string
	channelID          string
	blockchainProvider blockchainClientProvider
	jsonMarshal        func(v interface{}) ([]byte, error)
}

// NewFirstValidHandler returns a new, first valid handler
func NewFirstValidHandler(channelID string, cfg Config, blockchainProvider blockchainClientProvider) *FirstValid {
	return &FirstValid{
		Config:             cfg,
		path:               fmt.Sprintf("%s/firstValid", cfg.BasePath),
		channelID:          channelID,
		blockchainProvider: blockchainProvider,
		jsonMarshal:        json.Marshal,
	}
}

// Path returns the context path
func (h *FirstValid) Path() string {
	return h.path
}

// Method returns the HTTP method
func (h *FirstValid) Method() string {
	return http.MethodPost
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
		logger.Errorf("[%s] No valid transactions found: %s", h.channelID)

		return nil, httpserver.NotFoundError
	}

	logger.Debugf("[%s] Returning: %+v", h.channelID, resp)

	return &resp.Transactions[0], nil
}

func (h *FirstValid) blockchainClient() (bcclient.Blockchain, error) {
	bcClient, err := h.blockchainProvider.ForChannel(h.channelID)
	if err != nil {
		logger.Errorf("[%s] Failed to get blockchain client: %s", h.channelID, err)

		return nil, httpserver.ServerError
	}

	return bcClient, nil
}
