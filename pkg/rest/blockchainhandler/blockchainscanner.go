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
	channelID     string
	maxTxns       int
	sinceBlockNum uint64
	sinceTxnNum   uint64
	bcClient      bcclient.Blockchain
}

func newBlockchainScanner(channelID string, sinceBlockNum, sinceTxnNum uint64, maxTxns int, bcClient bcclient.Blockchain) *blockchainScanner {
	return &blockchainScanner{
		channelID:     channelID,
		maxTxns:       maxTxns,
		sinceBlockNum: sinceBlockNum,
		sinceTxnNum:   sinceTxnNum,
		bcClient:      bcClient,
	}
}

func (h *blockchainScanner) scan() (*TransactionsResponse, error) {
	bcInfo, err := h.bcClient.GetBlockchainInfo()
	if err != nil {
		return nil, errors.WithMessage(err, "failed to get blockchain info")
	}

	resp := &TransactionsResponse{}

	sinceTxnNum := h.sinceTxnNum

	for blockNum := h.sinceBlockNum; blockNum < bcInfo.Height; blockNum++ {
		txns, err := h.getSidetreeTransactions(blockNum, sinceTxnNum, h.maxTxns-len(resp.Transactions))
		if err != nil {
			if strings.Contains(err.Error(), errReachedMaxTxns.Error()) {
				resp.Transactions = append(resp.Transactions, txns...)
				resp.More = true

				logger.Debugf("[%s] Returning %d of more transactions: %+v", h.channelID, len(resp.Transactions), resp)

				return resp, nil
			}

			return nil, errors.WithMessagef(err, "failed to get transactions in block %d", blockNum)
		}

		// Reset sinceTxnNum to 0 since it only applies to the first block
		sinceTxnNum = 0
		resp.Transactions = append(resp.Transactions, txns...)
	}

	logger.Debugf("[%s] Returning %d transactions: %+v", h.channelID, len(resp.Transactions), resp)

	return resp, nil
}

func (h *blockchainScanner) getSidetreeTransactions(blockNum uint64, sinceTxnNum uint64, maxTxns int) ([]Transaction, error) {
	block, err := h.bcClient.GetBlockByNumber(blockNum)
	if err != nil {
		return nil, err
	}

	return newBlockScanner(h.channelID, block, sinceTxnNum, maxTxns).scan()
}
