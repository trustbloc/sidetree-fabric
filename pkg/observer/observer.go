/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package observer

import (
	"encoding/json"
	"strings"
	"sync/atomic"
	"time"

	cb "github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric/common/flogging"
	gossipapi "github.com/hyperledger/fabric/extensions/gossip/api"
	"github.com/pkg/errors"
	"github.com/trustbloc/fabric-peer-ext/pkg/common/blockvisitor"
	"github.com/trustbloc/sidetree-core-go/pkg/api/txn"
	"github.com/trustbloc/sidetree-core-go/pkg/observer"
	"github.com/trustbloc/sidetree-core-go/pkg/txnhandler"

	"github.com/trustbloc/sidetree-fabric/pkg/client"
	"github.com/trustbloc/sidetree-fabric/pkg/common/transienterr"
	"github.com/trustbloc/sidetree-fabric/pkg/config"
	ctxcommon "github.com/trustbloc/sidetree-fabric/pkg/context/common"
	"github.com/trustbloc/sidetree-fabric/pkg/observer/common"
	"github.com/trustbloc/sidetree-fabric/pkg/observer/lease"
	"github.com/trustbloc/sidetree-fabric/pkg/observer/operationfilter"
)

var logger = flogging.MustGetLogger("sidetree_observer")

const (
	defaultMonitorPeriod = 10 * time.Second
	defaultMaxAttempts   = 3
)

// Metadata contains meta-data for the document store
type Metadata struct {
	LastBlockProcessed uint64
	LastTxNumProcessed int64
	LeaseOwner         string
	FailedAttempts     int
	LastErrorCode      transienterr.Code
}

func newMetadata(leaseOwner string, lastBlockProcessed uint64) *Metadata {
	return &Metadata{
		LeaseOwner:         leaseOwner,
		LastBlockProcessed: lastBlockProcessed,
		LastTxNumProcessed: -1,
	}
}

// ClientProviders contains the providers for off-ledger, DCAS, and blockchain clients
type ClientProviders struct {
	OffLedger  common.OffLedgerClientProvider
	DCAS       common.DCASClientProvider
	Blockchain common.BlockchainClientProvider
	Gossip     common.GossipProvider
}

type metadataStore interface {
	Get() (*Metadata, error)
	Put(*Metadata) error
}

// Observer reads blocks from the ledger looking for Sidetree anchor writes and persists the document operations to the document store.
type Observer struct {
	leaseProvider  *lease.Provider
	channelID      string
	period         time.Duration
	maxAttempts    int
	done           chan struct{}
	txnProcessor   *observer.TxnProcessor
	blockchain     common.BlockchainClientProvider
	metadataStore  metadataStore
	processStarted uint32
	txnChan        <-chan gossipapi.TxMetadata
	processChan    chan struct{}
}

type peerConfig interface {
	PeerID() string
	MSPID() string
}

// New returns a new Observer
func New(channelID string, peerCfg peerConfig, observerCfg config.Observer, dcasCfg config.DCAS, clientProviders *ClientProviders, opStoreProvider ctxcommon.OperationStoreProvider, txnChan <-chan gossipapi.TxMetadata, pcp ctxcommon.ProtocolClientProvider) *Observer {
	period := observerCfg.Period
	if period == 0 {
		period = defaultMonitorPeriod
	}

	maxAttempts := observerCfg.MaxAttempts
	if maxAttempts == 0 {
		maxAttempts = defaultMaxAttempts
	}

	m := &Observer{
		channelID:     channelID,
		period:        period,
		maxAttempts:   maxAttempts,
		blockchain:    clientProviders.Blockchain,
		metadataStore: NewMetaDataStore(channelID, peerCfg, observerCfg.MetaDataChaincodeName, clientProviders.OffLedger),
		leaseProvider: lease.NewProvider(channelID, clientProviders.Gossip.GetGossipService()),
		txnProcessor: observer.NewTxnProcessor(
			&observer.Providers{
				TxnOpsProvider:   txnhandler.NewOperationProvider(NewSidetreeDCASReader(channelID, dcasCfg, clientProviders.DCAS), pcp),
				OpStoreProvider:  asObserverStoreProvider(opStoreProvider),
				OpFilterProvider: operationfilter.NewProvider(channelID, opStoreProvider),
			},
		),
		txnChan:     txnChan,
		done:        make(chan struct{}, 1),
		processChan: make(chan struct{}),
	}

	return m
}

// Start starts the Observer
func (m *Observer) Start() error {
	logger.Infof("[%s] Starting Observer", m.channelID)

	go m.run()
	go m.listen()

	return nil
}

// Stop stops the Observer
func (m *Observer) Stop() {
	logger.Infof("[%s] Stopping Observer", m.channelID)

	m.done <- struct{}{}
}

func (m *Observer) listen() {
	logger.Infof("[%s] Starting Observer with a period of %s", m.channelID, m.period)

	ticker := time.NewTicker(m.period)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			logger.Debugf("[%s] Triggering scheduled processing of Sidetree transactions", m.channelID)
			m.trigger()
		case txn := <-m.txnChan:
			logger.Debugf("[%s] Got notification about a Sidetree transaction in block %d and txnNum %d - triggering processing", m.channelID, txn.BlockNum, txn.TxNum)
			m.trigger()
		case <-m.done:
			logger.Infof("[%s] Exiting Observer", m.channelID)
			close(m.processChan)
			return
		}
	}
}

func (m *Observer) trigger() {
	if m.started() {
		logger.Debugf("[%s] No need to trigger the processor to start since it's already running", m.channelID)
		return
	}

	select {
	case m.processChan <- struct{}{}:
		logger.Debugf("[%s] Triggered processor for Sidetree transactions", m.channelID)
	default:
		logger.Debugf("[%s] Unable to trigger processor since the notification channel is busy", m.channelID)
	}
}

func (m *Observer) run() {
	logger.Infof("[%s] Listening for triggers", m.channelID)

	for range m.processChan {
		logger.Debugf("[%s] Got notification to process blocks for Sidetree transactions", m.channelID)
		m.process()
	}

	logger.Infof("[%s] ... stopped listening for triggers", m.channelID)
}

func (m *Observer) process() {
	if !m.processorStarting() {
		logger.Debugf("[%s] Processor already running", m.channelID)
		return
	}

	defer m.processorStopped()

	bcInfo, err := m.getBlockchainInfo()
	if err != nil {
		logger.Warnf("[%s] Error getting blockchain info: %s", m.channelID, err)
		return
	}

	metadata, err := m.getMetadata()
	if err != nil {
		logger.Warnf("[%s] Error getting metadata: %s", m.channelID, err)
		return
	}

	if ok := m.checkLeaseOwner(metadata); !ok {
		return
	}

	toBlockNum := bcInfo.Height - 1

	var fromBlockNum uint64
	if metadata.LastErrorCode != "" {
		// Got an error processing this block - need to retry the same block
		logger.Infof("[%s] Last processing of block:txNum [%d:%d] failed with error code [%s]. Will reprocess block [%d] starting at txNum [%d] - Attempt #%d", m.channelID, metadata.LastBlockProcessed, metadata.LastTxNumProcessed, metadata.LastErrorCode, metadata.LastBlockProcessed, metadata.LastTxNumProcessed, metadata.FailedAttempts+1)
		fromBlockNum = metadata.LastBlockProcessed
	} else {
		// The last block was successfully processed - start at the next block
		fromBlockNum = metadata.LastBlockProcessed + 1
		metadata.LastTxNumProcessed = -1
	}

	if fromBlockNum <= toBlockNum {
		logger.Debugf("[%s] Processing from block:txNum [%d:%d] to block [%d]", m.channelID, fromBlockNum, metadata.LastTxNumProcessed, toBlockNum)
	} else {
		logger.Debugf("[%s] No blocks to process. Last block processed: [%d]", m.channelID, metadata.LastBlockProcessed)
	}

	for bNum := fromBlockNum; bNum <= toBlockNum; bNum++ {
		err := m.processBlock(bNum, bNum == toBlockNum, metadata)

		m.putMetadata(metadata)

		if err != nil {
			logger.Errorf("[%s] Error processing block [%d]: %s", m.channelID, bNum, err)

			return
		}
	}
}

// checkLeaseOwner checks if the existing lease is still valid and returns true if this
// peer should start processing.
func (m *Observer) checkLeaseOwner(metadata *Metadata) bool {
	currentLease := m.leaseProvider.GetLease(metadata.LeaseOwner)

	// Check if the lease is still valid
	if currentLease.IsValid() {
		if !currentLease.IsLocalPeerOwner() {
			logger.Debugf("[%s] Not processing block %d since [%s] owns the lease", m.channelID, metadata.LastBlockProcessed+1, currentLease.Owner())
			return false
		}

		return true
	}

	// Create a new lease for the next block
	newLease := m.leaseProvider.CreateLease(metadata.LastBlockProcessed + 1)
	if !newLease.IsLocalPeerOwner() {
		logger.Debugf("[%s] Not processing block %d since [%s] is the new lease owner", m.channelID, metadata.LastBlockProcessed+1, newLease.Owner())
		return false
	}

	logger.Debugf("[%s] I am replacing [%s] as the new lease owner for block %d", m.channelID, currentLease.Owner(), metadata.LastBlockProcessed+1)

	// Save the metadata with the new lease owner and wait until the next period before processing it.
	// This avoids race conditions where multiple peers are trying to do the same thing.

	metadata.LeaseOwner = newLease.Owner()

	m.putMetadata(metadata)

	return false
}

func (m *Observer) processBlock(bNum uint64, changeLeaseOwner bool, metadata *Metadata) error {
	block, err := m.getBlockByNumber(bNum)
	if err != nil {
		return transienterr.New(errors.WithMessagef(err, "error getting block %d", bNum), transienterr.CodeDB)
	}

	logger.Debugf("[%s] Processing block [%d]", m.channelID, bNum)

	err = blockvisitor.New(m.channelID,
		blockvisitor.WithWriteHandler(m.writeHandler(metadata)),
		blockvisitor.WithErrorHandler(m.errorHandler(metadata))).Visit(block)
	if err != nil {
		return err
	}

	if changeLeaseOwner {
		metadata.LeaseOwner = m.leaseProvider.CreateLease(bNum + 1).Owner()

		logger.Debugf("[%s] Done processing blocks. Lease owner for the next block [%d] is [%s]", m.channelID, bNum+1, metadata.LeaseOwner)
	}

	metadata.LastBlockProcessed = bNum
	metadata.LastTxNumProcessed = -1
	metadata.LastErrorCode = ""
	metadata.FailedAttempts = 0

	return nil
}

func (m *Observer) writeHandler(metadata *Metadata) blockvisitor.WriteHandler {
	return func(w *blockvisitor.Write) error {
		if int64(w.TxNum) < metadata.LastTxNumProcessed {
			logger.Debugf("[%s] Ignoring write to key [%s] since block:txNum [%d:%d] has already been processed", m.channelID, w.Write.Key, w.BlockNum, w.TxNum)
			return nil
		}

		metadata.LastBlockProcessed = w.BlockNum
		metadata.LastTxNumProcessed = int64(w.TxNum)

		if !strings.HasPrefix(w.Write.Key, common.AnchorPrefix) {
			logger.Debugf("[%s] Ignoring write to namespace [%s] in block [%d] and TxNum [%d] since the key doesn't have the anchor address prefix [%s]", m.channelID, w.Namespace, w.BlockNum, w.TxNum, common.AnchorPrefix)

			return nil
		}

		var txnInfo common.TxnInfo
		if err := json.Unmarshal(w.Write.Value, &txnInfo); err != nil {
			return errors.WithMessagef(err, "unmarshal transaction info error for anchor [%s] in block [%d] and TxNum [%d]", w.Write.Key, w.BlockNum, w.TxNum)
		}

		logger.Debugf("[%s] Handling write to anchor [%s] in block [%d] and TxNum [%d] on attempt #%d", m.channelID, w.Write.Value, w.BlockNum, w.TxNum, metadata.FailedAttempts+1)

		sidetreeTxn := txn.SidetreeTxn{
			TransactionTime:   w.BlockNum,
			TransactionNumber: w.TxNum,
			AnchorString:      txnInfo.AnchorString,
			Namespace:         txnInfo.Namespace,
		}

		if err := m.txnProcessor.Process(sidetreeTxn); err != nil {
			return errors.WithMessagef(err, "error processing Txn for anchor [%s] in block [%d] and TxNum [%d]", w.Write.Key, w.BlockNum, w.TxNum)
		}

		return nil
	}
}

func (m *Observer) errorHandler(metadata *Metadata) blockvisitor.ErrorHandler {
	return func(err error, ctx *blockvisitor.Context) error {
		if ctx.Category == blockvisitor.UnmarshalErr {
			logger.Errorf("[%s] Ignoring persistent error in block:txNum [%d:%d]: %s. Context: %s", m.channelID, ctx.BlockNum, ctx.TxNum, err, ctx)

			metadata.FailedAttempts = 0
			metadata.LastErrorCode = ""

			return nil
		}

		if !transienterr.Is(err) {
			logger.Errorf("[%s] Ignoring persistent error in block:txNum [%d:%d]: %s. Context: %s", m.channelID, ctx.BlockNum, ctx.TxNum, err, ctx)

			metadata.FailedAttempts = 0
			metadata.LastErrorCode = ""

			return nil
		}

		code := transienterr.GetCode(err)

		logger.Debugf("[%s] Got error processing block:txNum [%d:%d] after attempt #%d: %s - Code: %s. Context: %s", m.channelID, ctx.BlockNum, ctx.TxNum, metadata.FailedAttempts+1, err, code, ctx)

		if metadata.LastErrorCode == code && metadata.LastTxNumProcessed == int64(ctx.TxNum) {
			metadata.FailedAttempts++

			if metadata.FailedAttempts > m.maxAttempts {
				logger.Errorf("[%s] Giving up processing block:txNum [%d:%d] after %d failed attempts on error: %s. Context: %s", m.channelID, ctx.BlockNum, ctx.TxNum, metadata.FailedAttempts+1, err, ctx)

				metadata.FailedAttempts = 0
				metadata.LastErrorCode = ""

				return nil
			}

			logger.Warnf("[%s] Got same error as before processing block:txNum [%d:%d]: %s - Code: %s. Increasing failed attempts to %d", m.channelID, ctx.BlockNum, ctx.TxNum, err, code, metadata.FailedAttempts)
		} else {
			// Reset FailedAttempts since the error/txNum is different this time
			metadata.FailedAttempts = 1

			logger.Warnf("[%s] Got new error processing block:txNum [%d:%d]: %s - Code: %s", m.channelID, ctx.BlockNum, ctx.TxNum, err, code)
		}

		metadata.LastErrorCode = code

		return err
	}
}

func (m *Observer) getBlockchainInfo() (*cb.BlockchainInfo, error) {
	bcClient, err := m.blockchainClient()
	if err != nil {
		return nil, err
	}
	block, err := bcClient.GetBlockchainInfo()
	if err != nil {
		return nil, errors.WithMessage(err, "failed to get blockchain info")
	}
	return block, nil
}

func (m *Observer) getBlockByNumber(bNum uint64) (*cb.Block, error) {
	bcClient, err := m.blockchainClient()
	if err != nil {
		return nil, err
	}
	block, err := bcClient.GetBlockByNumber(bNum)
	if err != nil {
		return nil, errors.WithMessagef(err, "failed to get block number [%d]", bNum)
	}
	return block, nil
}

func (m *Observer) blockchainClient() (client.Blockchain, error) {
	return m.blockchain.ForChannel(m.channelID)
}

func (m *Observer) getMetadata() (*Metadata, error) {
	metadata, err := m.metadataStore.Get()
	if err != nil {
		if err == errMetaDataNotFound {
			return newMetadata(m.leaseProvider.CreateLease(1).Owner(), 0), nil
		}

		return nil, err
	}

	return metadata, nil
}

func (m *Observer) putMetadata(metadata *Metadata) {
	logger.Debugf("[%s] Updating metadata: %+v", m.channelID, metadata)

	if err := m.metadataStore.Put(metadata); err != nil {
		logger.Errorf("[%s] Error saving metadata: %s", m.channelID, err)
	}
}

func (m *Observer) started() bool {
	return atomic.LoadUint32(&m.processStarted) == 1
}

func (m *Observer) processorStarting() bool {
	return atomic.CompareAndSwapUint32(&m.processStarted, 0, 1)
}

func (m *Observer) processorStopped() {
	atomic.CompareAndSwapUint32(&m.processStarted, 1, 0)
}

type observerStorePovider struct {
	opStoreProvider ctxcommon.OperationStoreProvider
}

func asObserverStoreProvider(p ctxcommon.OperationStoreProvider) *observerStorePovider {
	return &observerStorePovider{opStoreProvider: p}
}

func (p *observerStorePovider) ForNamespace(namespace string) (observer.OperationStore, error) {
	s, err := p.opStoreProvider.ForNamespace(namespace)
	if err != nil {
		return nil, err
	}

	return s, nil
}
