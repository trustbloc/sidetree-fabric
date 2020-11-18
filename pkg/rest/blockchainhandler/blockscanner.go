/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package blockchainhandler

import (
	"encoding/base64"
	"encoding/json"
	"strings"

	cb "github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric/protoutil"
	"github.com/pkg/errors"
	"github.com/trustbloc/fabric-peer-ext/pkg/common/blockvisitor"

	bcclient "github.com/trustbloc/sidetree-fabric/pkg/client"
	"github.com/trustbloc/sidetree-fabric/pkg/observer/common"
)

var errReachedMaxTxns = errors.New("maximum transactions reached")

type blockScanner struct {
	channelID string
	bcClient  bcclient.Blockchain
}

func newBlockScanner(channelID string, bcClient bcclient.Blockchain) *blockScanner {
	return &blockScanner{
		channelID: channelID,
		bcClient:  bcClient,
	}
}

func (p *blockScanner) Scan(desc BlockDescriptor, maxTxns int) ([]Transaction, bool, error) {
	block, err := p.bcClient.GetBlockByNumber(desc.BlockNum())
	if err != nil {
		return nil, false, err
	}

	return newTxnBlockScanner(p.channelID, block, desc.TxnNum(), maxTxns).scan()
}

type txnBlockScanner struct {
	channelID           string
	block               *cb.Block
	sinceTxnNum         uint64
	maxTxns             int
	transactionTimeHash string
	transactions        []Transaction
}

func newTxnBlockScanner(channelID string, block *cb.Block, sinceTxnNum uint64, maxTxns int) *txnBlockScanner {
	return &txnBlockScanner{
		channelID:   channelID,
		block:       block,
		sinceTxnNum: sinceTxnNum,
		maxTxns:     maxTxns,
	}
}

func (h *txnBlockScanner) scan() ([]Transaction, bool, error) {
	blockHash := protoutil.BlockHeaderHash(h.block.Header)
	h.transactionTimeHash = base64.URLEncoding.EncodeToString(blockHash)

	visitor := blockvisitor.New(h.channelID,
		blockvisitor.WithWriteHandler(h.handleWrite),
		blockvisitor.WithErrorHandler(h.handleError),
	)

	err := visitor.Visit(h.block, nil)
	if err != nil {
		if errors.Cause(err) == errReachedMaxTxns {
			return h.transactions, true, nil
		}

		return nil, false, err
	}

	return h.transactions, false, nil
}

func (h *txnBlockScanner) handleWrite(w *blockvisitor.Write) error {
	if !strings.HasPrefix(w.Write.Key, common.AnchorPrefix) {
		logger.Debugf("[%s] Ignoring write to namespace [%s] in block [%d] and TxNum [%d] since the key doesn't have the anchor string prefix [%s]", h.channelID, w.Namespace, w.BlockNum, w.TxNum, common.AnchorPrefix)

		return nil
	}

	if w.TxNum < h.sinceTxnNum {
		logger.Debugf("[%s] Ignoring write in block [%d] and TxNum [%d] since the transaction number is less than the 'sinceTxnNum' %d", h.channelID, w.BlockNum, w.TxNum, h.sinceTxnNum)

		return nil
	}

	if len(h.transactions) >= h.maxTxns {
		return errReachedMaxTxns
	}

	anchorString, err := getAnchorString(w.Write.Value)
	if err != nil {
		return errors.WithMessagef(err, "failed to get anchor string [%s] in block [%d] and TxNum [%d]", w.Write.Key, w.BlockNum, w.TxNum)
	}

	txn := Transaction{
		TransactionNumber:   w.TxNum,
		TransactionTime:     w.BlockNum,
		TransactionTimeHash: h.transactionTimeHash,
		AnchorString:        anchorString,
	}

	logger.Debugf("[%s] Adding transaction %+v", h.channelID, txn)

	h.transactions = append(h.transactions, txn)

	return nil
}

func (h *txnBlockScanner) handleError(err error, ctx *blockvisitor.Context) error {
	if err == errReachedMaxTxns {
		logger.Debugf("[%s] Reached the maximum number of transactions", h.channelID)

		return err
	}

	logger.Errorf("[%s] Error processing block: %s. Context: %s. Block will be ignored.", h.channelID, err, ctx)
	return nil
}

func getAnchorString(value []byte) (string, error) {
	var txnInfo common.TxnInfo
	if err := json.Unmarshal(value, &txnInfo); err != nil {
		return "", err
	}

	return txnInfo.AnchorString, nil
}
