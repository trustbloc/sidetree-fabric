/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package blockchainhandler

import (
	"strings"

	"github.com/pkg/errors"
	bcclient "github.com/trustbloc/sidetree-fabric/pkg/client"
)

type blockchainScanner struct {
	channelID string
	maxTxns   int
	bcClient  bcclient.Blockchain
}

func newBlockchainScanner(channelID string, maxTxns int, bcClient bcclient.Blockchain) *blockchainScanner {
	return &blockchainScanner{
		channelID: channelID,
		maxTxns:   maxTxns,
		bcClient:  bcClient,
	}
}

func (h *blockchainScanner) scan() (*TransactionsResponse, error) {
	bcInfo, err := h.bcClient.GetBlockchainInfo()
	if err != nil {
		return nil, errors.WithMessage(err, "failed to get blockchain info")
	}

	resp := &TransactionsResponse{}

	// Start at block 1 since block 0 is the genesis block
	for blockNum := uint64(1); blockNum < bcInfo.Height; blockNum++ {
		txns, err := h.getSidetreeTransactions(blockNum, h.maxTxns-len(resp.Transactions))
		if err != nil {
			if strings.Contains(err.Error(), errReachedMaxTxns.Error()) {
				resp.Transactions = append(resp.Transactions, txns...)
				resp.More = true

				logger.Debugf("[%s] Returning %d of more transactions: %+v", h.channelID, len(resp.Transactions), resp)

				return resp, nil
			}

			return nil, errors.WithMessagef(err, "failed to get transactions in block %d", blockNum)
		}

		resp.Transactions = append(resp.Transactions, txns...)
	}

	logger.Debugf("[%s] Returning %d transactions: %+v", h.channelID, len(resp.Transactions), resp)

	return resp, nil
}

func (h *blockchainScanner) getSidetreeTransactions(blockNum uint64, maxTxns int) ([]Transaction, error) {
	block, err := h.bcClient.GetBlockByNumber(blockNum)
	if err != nil {
		return nil, err
	}

	return newBlockScanner(h.channelID, block, maxTxns).scan()
}
