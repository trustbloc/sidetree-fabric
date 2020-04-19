/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package blockchainhandler

import (
	"encoding/base64"
	"strings"

	cb "github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric/protoutil"
	"github.com/pkg/errors"
	"github.com/trustbloc/fabric-peer-ext/pkg/common/blockvisitor"

	"github.com/trustbloc/sidetree-fabric/pkg/observer/common"
)

var errReachedMaxTxns = errors.New("maximum transactions reached")

type blockScanner struct {
	channelID           string
	block               *cb.Block
	maxTxns             int
	transactionTimeHash string
	transactions        []Transaction
}

func newBlockScanner(channelID string, block *cb.Block, maxTxns int) *blockScanner {
	return &blockScanner{
		channelID: channelID,
		block:     block,
		maxTxns:   maxTxns,
	}
}

func (h *blockScanner) scan() ([]Transaction, error) {
	blockHash := protoutil.BlockHeaderHash(h.block.Header)
	h.transactionTimeHash = base64.URLEncoding.EncodeToString(blockHash)

	visitor := blockvisitor.New(h.channelID,
		blockvisitor.WithWriteHandler(h.handleWrite),
		blockvisitor.WithErrorHandler(h.handleError),
	)

	return h.transactions, visitor.Visit(h.block)
}

func (h *blockScanner) handleWrite(w *blockvisitor.Write) error {
	if !strings.HasPrefix(w.Write.Key, common.AnchorAddrPrefix) {
		logger.Debugf("[%s] Ignoring write to namespace [%s] in block [%d] and TxNum [%d] since the key doesn't have the anchor address prefix [%s]", h.channelID, w.Namespace, w.BlockNum, w.TxNum, common.AnchorAddrPrefix)

		return nil
	}

	if len(h.transactions) >= h.maxTxns {
		return errReachedMaxTxns
	}

	txn := Transaction{
		TransactionNumber:   w.TxNum,
		TransactionTime:     w.BlockNum,
		TransactionTimeHash: h.transactionTimeHash,
		AnchorString:        string(w.Write.Value),
	}

	logger.Debugf("[%s] Adding transaction %+v", h.channelID, txn)

	h.transactions = append(h.transactions, txn)

	return nil
}

func (h *blockScanner) handleError(err error, ctx *blockvisitor.Context) error {
	if err == errReachedMaxTxns {
		logger.Debugf("[%s] Reached the maximum number of transactions", h.channelID)

		return err
	}

	logger.Errorf("[%s] Error processing block: %s. Context: %s. Block will be ignored.", h.channelID, err, ctx)
	return nil
}
