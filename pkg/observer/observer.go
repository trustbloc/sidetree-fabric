/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package observer

import (
	"strings"
	"sync/atomic"
	"time"

	cb "github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric/common/flogging"
	gossipapi "github.com/hyperledger/fabric/extensions/gossip/api"
	"github.com/pkg/errors"
	"github.com/trustbloc/fabric-peer-ext/pkg/common/blockvisitor"
	"github.com/trustbloc/sidetree-core-go/pkg/observer"

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
)

// Metadata contains meta-data for the document store
type Metadata struct {
	LastBlockProcessed uint64
	LeaseOwner         string
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
	blockVisitor   *blockvisitor.Visitor
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
func New(channelID string, peerCfg peerConfig, observerCfg config.Observer, dcasCfg config.DCAS, clientProviders *ClientProviders, opStoreProvider ctxcommon.OperationStoreProvider, txnChan <-chan gossipapi.TxMetadata) *Observer {
	period := observerCfg.Period
	if period == 0 {
		period = defaultMonitorPeriod
	}

	m := &Observer{
		channelID:     channelID,
		period:        period,
		blockchain:    clientProviders.Blockchain,
		metadataStore: NewMetaDataStore(channelID, peerCfg, observerCfg.MetaDataChaincodeName, clientProviders.OffLedger),
		leaseProvider: lease.NewProvider(channelID, clientProviders.Gossip.GetGossipService()),
		txnProcessor: observer.NewTxnProcessor(
			&observer.Providers{
				DCASClient:       NewSidetreeDCASReader(channelID, dcasCfg, clientProviders.DCAS),
				OpStoreProvider:  asObserverStoreProvider(opStoreProvider),
				OpFilterProvider: operationfilter.NewProvider(channelID, opStoreProvider),
			},
		),
		txnChan:     txnChan,
		done:        make(chan struct{}, 1),
		processChan: make(chan struct{}),
	}

	m.blockVisitor = blockvisitor.New(channelID,
		blockvisitor.WithWriteHandler(m.handleWrite),
		blockvisitor.WithErrorHandler(m.handleError),
	)

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

	lastBlockNum := metadata.LastBlockProcessed
	toBlockNum := bcInfo.Height - 1

	if lastBlockNum < toBlockNum {
		logger.Debugf("[%s] Processing blocks [%d:%d]", m.channelID, lastBlockNum+1, toBlockNum)
	} else {
		logger.Debugf("[%s] No blocks to process. Last block processed: [%d]", m.channelID, lastBlockNum)
	}

	for bNum := lastBlockNum + 1; bNum <= toBlockNum; bNum++ {
		logger.Debugf("[%s] Processing block [%d]", m.channelID, bNum)

		if err = m.processBlock(bNum); err != nil {
			logger.Errorf("[%s] Error processing block [%d]: %s", m.channelID, bNum, err)
			return
		}

		metadata.LastBlockProcessed = bNum

		if bNum == toBlockNum {
			metadata.LeaseOwner = m.leaseProvider.CreateLease(bNum + 1).Owner()

			logger.Debugf("[%s] Done processing blocks [%d:%d]. Lease owner for the next block is [%s]", m.channelID, lastBlockNum+1, toBlockNum, metadata.LeaseOwner)
		}

		err = m.putMetadata(metadata)
		if err != nil {
			logger.Errorf("[%s] Error saving metadata for block [%d]: %s", m.channelID, bNum, err)
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

	if err := m.putMetadata(metadata); err != nil {
		logger.Errorf("[%s] Error saving metadata: %s", m.channelID, err)
	}

	return false
}

func (m *Observer) processBlock(bNum uint64) error {
	var block *cb.Block
	var err error

	if block, err = m.getBlockByNumber(bNum); err != nil {
		return errors.WithMessagef(err, "error getting block [%d]", bNum)
	}

	return m.blockVisitor.Visit(block)
}

func (m *Observer) handleWrite(w *blockvisitor.Write) error {
	if !strings.HasPrefix(w.Write.Key, common.AnchorAddrPrefix) {
		logger.Debugf("[%s] Ignoring write to namespace [%s] in block [%d] and TxNum [%d] since the key doesn't have the anchor address prefix [%s]", m.channelID, w.Namespace, w.BlockNum, w.TxNum, common.AnchorAddrPrefix)
		return nil
	}

	logger.Debugf("[%s] Handling write to anchor [%s] in block [%d] and TxNum [%d]", m.channelID, w.Write.Value, w.BlockNum, w.TxNum)

	sidetreeTxn := observer.SidetreeTxn{
		TransactionTime:   w.BlockNum,
		TransactionNumber: w.TxNum,
		AnchorAddress:     string(w.Write.Value),
	}

	if err := m.txnProcessor.Process(sidetreeTxn); err != nil {
		return errors.WithMessagef(err, "error processing Txn for anchor [%s] in block [%d] and TxNum [%d]", w.Write.Key, w.BlockNum, w.TxNum)
	}

	return nil
}

func (m *Observer) handleError(err error, ctx *blockvisitor.Context) error {
	if ctx.Category == blockvisitor.UnmarshalErr {
		logger.Errorf("[%s] Ignoring persistent error: %s. Context: %s", m.channelID, err, ctx)
		return nil
	}

	if !transienterr.Is(err) {
		logger.Errorf("[%s] Ignoring persistent error: %s. Context: %s", m.channelID, err, ctx)
		return nil
	}

	logger.Warnf("[%s] Will retry on transient error [%s] in %s. Context: %s", m.channelID, err, m.period, ctx)

	return err
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
			return &Metadata{
				LastBlockProcessed: 0,
				LeaseOwner:         m.leaseProvider.CreateLease(1).Owner(),
			}, nil
		}

		return nil, err
	}

	return metadata, nil
}

func (m *Observer) putMetadata(metadata *Metadata) error {
	logger.Debugf("[%s] Updating metadata: %+v", m.channelID, metadata)

	if err := m.metadataStore.Put(metadata); err != nil {
		return errors.WithMessage(err, "error persisting metadata")
	}

	return nil
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
