/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package observer

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync/atomic"

	cb "github.com/hyperledger/fabric-protos-go/common"
	"github.com/pkg/errors"
	"github.com/trustbloc/fabric-peer-ext/pkg/common/blockvisitor"
	"github.com/trustbloc/sidetree-core-go/pkg/api/protocol"
	"github.com/trustbloc/sidetree-core-go/pkg/api/txn"

	"github.com/trustbloc/sidetree-fabric/pkg/client"
	"github.com/trustbloc/sidetree-fabric/pkg/common/transienterr"
	ctxcommon "github.com/trustbloc/sidetree-fabric/pkg/context/common"
	"github.com/trustbloc/sidetree-fabric/pkg/observer/common"
)

type behavior struct {
	processTxn       func(sidetreeTxn *txn.SidetreeTxn, pv protocol.Version) error
	getMetadata      func() (Metadata, bool, error)
	putMetadata      func(metadata *Metadata)
	changeLeaseOwner func(metadata *Metadata, blockNum uint64)
}

type processor struct {
	*behavior
	id             string // Used only for logging
	channelID      string
	pcp            ctxcommon.ProtocolClientProvider
	maxAttempts    int
	blockchain     common.BlockchainClientProvider
	processStarted uint32
}

func newProcessor(
	pType, channelID string, behavior *behavior,
	maxAttempts int, pcp ctxcommon.ProtocolClientProvider,
	blockchain common.BlockchainClientProvider) *processor {

	return &processor{
		id:          fmt.Sprintf("%s:%s", channelID, pType),
		behavior:    behavior,
		channelID:   channelID,
		pcp:         pcp,
		maxAttempts: maxAttempts,
		blockchain:  blockchain,
	}
}

func (p *processor) ProcessBlockchain() {
	if !p.processorStarting() {
		logger.Debugf("[%s] Processor already running", p.id)
		return
	}

	defer p.processorStopped()

	logger.Debugf("[%s] Processing started.", p.id)

	bcInfo, err := p.getBlockchainInfo()
	if err != nil {
		logger.Warnf("[%s] Error getting blockchain info: %s", p.id, err)
		return
	}

	md, ok, err := p.getMetadata()
	if err != nil {
		logger.Warnf("[%s] Error getting metadata: %s", p.id, err)
		return
	}

	if !ok {
		logger.Debugf("[%s] Not processing since I'm not the lease owner", p.id)

		return
	}

	metadata := &md

	toBlockNum := bcInfo.Height - 1

	var fromBlockNum uint64
	if metadata.LastErrorCode != "" {
		// Got an error processing this block - need to retry the same block
		logger.Infof("[%s] Last processing of block:txNum [%d:%d] failed with error code [%s]. Will reprocess block [%d] starting at txNum [%d] - Attempt #%d",
			p.id, metadata.LastBlockProcessed, metadata.LastTxNumProcessed, metadata.LastErrorCode,
			metadata.LastBlockProcessed, metadata.LastTxNumProcessed, metadata.FailedAttempts+1)

		fromBlockNum = metadata.LastBlockProcessed
	} else {
		// The last block was successfully processed - start at the next block
		fromBlockNum = metadata.LastBlockProcessed + 1
		metadata.LastTxNumProcessed = -1
	}

	if fromBlockNum <= toBlockNum {
		logger.Debugf("[%s] Processing from block:txNum [%d:%d] to block [%d]", p.id, fromBlockNum, metadata.LastTxNumProcessed, toBlockNum)
	} else {
		logger.Debugf("[%s] No blocks to process. Last block processed: [%d]", p.id, metadata.LastBlockProcessed)
	}

	p.processBlocks(fromBlockNum, toBlockNum, metadata)
}

func (p *processor) processBlocks(fromBlockNum, toBlockNum uint64, metadata *Metadata) {
	for bNum := fromBlockNum; bNum <= toBlockNum; bNum++ {
		err := p.processBlock(bNum, bNum == toBlockNum, metadata)

		p.putMetadata(metadata)

		if err != nil {
			logger.Errorf("[%s] Error processing block [%d]: %s", p.channelID, bNum, err)

			return
		}
	}
}

func (p *processor) processBlock(bNum uint64, changeLeaseOwner bool, metadata *Metadata) error {
	block, err := p.getBlockByNumber(bNum)
	if err != nil {
		return transienterr.New(errors.WithMessagef(err, "error getting block %d", bNum), transienterr.CodeDB)
	}

	logger.Debugf("[%s] Processing block [%d]", p.channelID, bNum)

	err = blockvisitor.New(p.channelID,
		blockvisitor.WithWriteHandler(p.writeHandler(metadata)),
		blockvisitor.WithErrorHandler(p.errorHandler(metadata))).Visit(block, nil)
	if err != nil {
		return err
	}

	if changeLeaseOwner {
		p.changeLeaseOwner(metadata, bNum+1)
	}

	metadata.LastBlockProcessed = bNum
	metadata.LastTxNumProcessed = -1
	metadata.LastErrorCode = ""
	metadata.FailedAttempts = 0

	return nil
}

func (p *processor) writeHandler(metadata *Metadata) blockvisitor.WriteHandler {
	return func(w *blockvisitor.Write) error {
		if int64(w.TxNum) < metadata.LastTxNumProcessed {
			logger.Debugf("[%s] Ignoring write to key [%s] since block:txNum [%d:%d] has already been processed", p.channelID, w.Write.Key, w.BlockNum, w.TxNum)
			return nil
		}

		metadata.LastBlockProcessed = w.BlockNum
		metadata.LastTxNumProcessed = int64(w.TxNum)

		if !strings.HasPrefix(w.Write.Key, common.AnchorPrefix) {
			logger.Debugf("[%s] Ignoring write to namespace [%s] in block [%d] and TxNum [%d] since the key doesn't have the anchor address prefix [%s]", p.channelID, w.Namespace, w.BlockNum, w.TxNum, common.AnchorPrefix)

			return nil
		}

		return p.processAnchor(w, metadata.FailedAttempts+1)
	}
}

func (p *processor) processAnchor(w *blockvisitor.Write, attemptNum int) error {
	sidetreeTxn, err := p.unmarshalTransaction(w)
	if err != nil {
		return err
	}

	logger.Debugf("[%s:%s] Handling write to anchor [%s] in block [%d] and TxNum [%d] on attemptNum #%d", p.channelID, sidetreeTxn.Namespace, w.Write.Key, w.BlockNum, w.TxNum, attemptNum)

	pv, err := p.getProtocolVersion(sidetreeTxn.Namespace, w.BlockNum)
	if err != nil {
		return errors.WithMessagef(err, "unable to get protocol version for namespace [%s] and block number [%d]", sidetreeTxn.Namespace, w.BlockNum)
	}

	err = p.processTxn(sidetreeTxn, pv)
	if err != nil {
		return errors.WithMessagef(err, "error processing Txn for anchor [%s] in block [%d] and TxNum [%d]", w.Write.Key, w.BlockNum, w.TxNum)
	}

	return nil
}

func (p *processor) getProtocolVersion(namespace string, blockNum uint64) (protocol.Version, error) {
	pc, err := p.pcp.ForNamespace(namespace)
	if err != nil {
		return nil, err
	}

	return pc.Get(blockNum)
}

func (p *processor) unmarshalTransaction(w *blockvisitor.Write) (*txn.SidetreeTxn, error) {
	var txnInfo common.TxnInfo
	if err := json.Unmarshal(w.Write.Value, &txnInfo); err != nil {
		return nil, errors.WithMessagef(err, "unmarshal transaction info error for anchor [%s] in block [%d] and TxNum [%d]", w.Write.Key, w.BlockNum, w.TxNum)
	}

	return &txn.SidetreeTxn{
		TransactionTime:     w.BlockNum,
		TransactionNumber:   w.TxNum,
		AnchorString:        txnInfo.AnchorString,
		Namespace:           txnInfo.Namespace,
		ProtocolGenesisTime: txnInfo.ProtocolGenesisTime,
	}, nil
}

func (p *processor) errorHandler(metadata *Metadata) blockvisitor.ErrorHandler {
	return func(err error, ctx *blockvisitor.Context) error {
		if ctx.Category == blockvisitor.UnmarshalErr {
			logger.Errorf("[%s] Ignoring persistent error in block:txNum [%d:%d]: %s. Context: %s", p.channelID, ctx.BlockNum, ctx.TxNum, err, ctx)

			metadata.FailedAttempts = 0
			metadata.LastErrorCode = ""

			return nil
		}

		if !transienterr.Is(err) {
			logger.Errorf("[%s] Ignoring persistent error in block:txNum [%d:%d]: %s. Context: %s", p.channelID, ctx.BlockNum, ctx.TxNum, err, ctx)

			metadata.FailedAttempts = 0
			metadata.LastErrorCode = ""

			return nil
		}

		code := transienterr.GetCode(err)

		logger.Debugf("[%s] Got error processing block:txNum [%d:%d] after attempt #%d: %s - Code: %s. Context: %s", p.channelID, ctx.BlockNum, ctx.TxNum, metadata.FailedAttempts+1, err, code, ctx)

		if metadata.LastErrorCode == code && metadata.LastTxNumProcessed == int64(ctx.TxNum) {
			metadata.FailedAttempts++

			if metadata.FailedAttempts > p.maxAttempts {
				logger.Errorf("[%s] Giving up processing block:txNum [%d:%d] after %d failed attempts on error: %s. Context: %s", p.channelID, ctx.BlockNum, ctx.TxNum, metadata.FailedAttempts+1, err, ctx)

				metadata.FailedAttempts = 0
				metadata.LastErrorCode = ""

				return nil
			}

			logger.Warnf("[%s] Got same error as before processing block:txNum [%d:%d]: %s - Code: %s. Increasing failed attempts to %d", p.channelID, ctx.BlockNum, ctx.TxNum, err, code, metadata.FailedAttempts)
		} else {
			// Reset FailedAttempts since the error/txNum is different this time
			metadata.FailedAttempts = 1

			logger.Warnf("[%s] Got new error processing block:txNum [%d:%d]: %s - Code: %s", p.channelID, ctx.BlockNum, ctx.TxNum, err, code)
		}

		metadata.LastErrorCode = code

		return err
	}
}

func (p *processor) getBlockchainInfo() (*cb.BlockchainInfo, error) {
	bcClient, err := p.blockchainClient()
	if err != nil {
		return nil, err
	}
	block, err := bcClient.GetBlockchainInfo()
	if err != nil {
		return nil, errors.WithMessage(err, "failed to get blockchain info")
	}
	return block, nil
}

func (p *processor) getBlockByNumber(bNum uint64) (*cb.Block, error) {
	bcClient, err := p.blockchainClient()
	if err != nil {
		return nil, err
	}
	block, err := bcClient.GetBlockByNumber(bNum)
	if err != nil {
		return nil, errors.WithMessagef(err, "failed to get block number [%d]", bNum)
	}
	return block, nil
}

func (p *processor) blockchainClient() (client.Blockchain, error) {
	return p.blockchain.ForChannel(p.channelID)
}

func (p *processor) processorStarting() bool {
	return atomic.CompareAndSwapUint32(&p.processStarted, 0, 1)
}

func (p *processor) processorStopped() {
	atomic.CompareAndSwapUint32(&p.processStarted, 1, 0)
}
