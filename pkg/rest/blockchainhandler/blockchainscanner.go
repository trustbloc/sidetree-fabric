/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package blockchainhandler

import (
	"github.com/pkg/errors"
)

// BlockDescriptor contains the block number and transaction number
type BlockDescriptor interface {
	BlockNum() uint64
	TxnNum() uint64
}

// BlockIterator iterates through a set of blocks and returns
// a block descriptor or false if there are no more blocks
type BlockIterator interface {
	Next() (BlockDescriptor, bool)
}

// BlockScanner scans a block and returns the Sidetree transactions for the block.
// The transactions are returned along with a boolean to indicate whether the maximum
// count has been reached.
type BlockScanner interface {
	Scan(desc BlockDescriptor, maxTxns int) ([]Transaction, bool, error)
}

func scanBlockchain(channelID string, maxTxns int, blockScanner BlockScanner, it BlockIterator) (*TransactionsResponse, error) {
	resp := &TransactionsResponse{}

	for {
		desc, ok := it.Next()
		if !ok {
			break
		}

		txns, reachedMax, err := blockScanner.Scan(desc, maxTxns-len(resp.Transactions))
		if err != nil {
			return nil, errors.WithMessagef(err, "failed to get transactions in block %d", desc.BlockNum())
		}

		resp.Transactions = append(resp.Transactions, txns...)

		if reachedMax {
			resp.More = true
			break
		}
	}

	logger.Debugf("[%s] Returning %d transactions: %+v", channelID, len(resp.Transactions), resp)

	return resp, nil
}
