/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package observer

import (
	"sync"
	"time"

	"github.com/hyperledger/fabric/common/flogging"
	gossipapi "github.com/hyperledger/fabric/extensions/gossip/api"
	"github.com/pkg/errors"
	"github.com/trustbloc/sidetree-core-go/pkg/api/protocol"
	"github.com/trustbloc/sidetree-core-go/pkg/api/txn"

	"github.com/trustbloc/sidetree-fabric/pkg/common/transienterr"
	"github.com/trustbloc/sidetree-fabric/pkg/config"
	ctxcommon "github.com/trustbloc/sidetree-fabric/pkg/context/common"
	"github.com/trustbloc/sidetree-fabric/pkg/context/doccache"
	"github.com/trustbloc/sidetree-fabric/pkg/observer/common"
	"github.com/trustbloc/sidetree-fabric/pkg/observer/lease"
	"github.com/trustbloc/sidetree-fabric/pkg/role"
)

var logger = flogging.MustGetLogger("sidetree_observer")

type cacheInvalidatorProvider interface {
	GetDocumentInvalidator(channelID, namespace string) (doccache.Invalidator, error)
}

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
	OffLedger                common.OffLedgerClientProvider
	Blockchain               common.BlockchainClientProvider
	Gossip                   common.GossipProvider
	CacheInvalidatorProvider cacheInvalidatorProvider
}

type metadataStore interface {
	Get() (*Metadata, error)
	Put(*Metadata) error
}

type blockchainProcessor interface {
	ProcessBlockchain()
}

// Observer reads blocks from the ledger looking for Sidetree anchor writes and persists the document operations to the document store.
type Observer struct {
	leaseProvider            *lease.Provider
	channelID                string
	period                   time.Duration
	done                     chan struct{}
	metadataStore            metadataStore
	txnChan                  <-chan gossipapi.TxMetadata
	processChan              chan struct{}
	cacheInvalidatorProvider cacheInvalidatorProvider
	cacheMetadata            Metadata
	mutex                    sync.RWMutex
	process                  func()
}

type peerConfig interface {
	PeerID() string
	MSPID() string
}

// New returns a new Observer
func New(channelID string, peerCfg peerConfig, observerCfg config.Observer, clientProviders *ClientProviders, txnChan <-chan gossipapi.TxMetadata, pcp ctxcommon.ProtocolClientProvider) *Observer {
	period := observerCfg.Period
	if period == 0 {
		period = defaultMonitorPeriod
	}

	m := &Observer{
		channelID:                channelID,
		period:                   period,
		metadataStore:            NewMetaDataStore(channelID, peerCfg, observerCfg.MetaDataChaincodeName, clientProviders.OffLedger),
		leaseProvider:            lease.NewProvider(channelID, clientProviders.Gossip.GetGossipService()),
		txnChan:                  txnChan,
		done:                     make(chan struct{}, 1),
		processChan:              make(chan struct{}),
		cacheInvalidatorProvider: clientProviders.CacheInvalidatorProvider,
	}

	m.process = m.createProcessor(observerCfg, pcp, clientProviders.Blockchain)

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

func (m *Observer) createProcessor(observerCfg config.Observer, pcp ctxcommon.ProtocolClientProvider, bcp common.BlockchainClientProvider) func() {
	maxAttempts := observerCfg.MaxAttempts
	if maxAttempts == 0 {
		maxAttempts = defaultMaxAttempts
	}

	var processors []blockchainProcessor

	if isObserver() {
		processors = append(processors, newProcessor("processor", m.channelID, m.processorBehavior(), maxAttempts, pcp, bcp))
	}

	if isResolver() {
		processors = append(processors, newProcessor("cache-invalidator", m.channelID, m.cacheInvalidatorBehavior(), maxAttempts, pcp, bcp))
	}

	return func() {
		for _, p := range processors {
			p.ProcessBlockchain()
		}
	}
}

// processorBehavior persists operations to the store
func (m *Observer) processorBehavior() *behavior {
	return &behavior{
		processTxn:       m.processTxn,
		getMetadata:      m.getMetadata,
		putMetadata:      m.putMetadata,
		changeLeaseOwner: m.changeLeaseOwner,
	}
}

// cacheInvalidatorBehavior invalidates the document cache for any updated documents
func (m *Observer) cacheInvalidatorBehavior() *behavior {
	return &behavior{
		processTxn:       m.processTxnForCache,
		getMetadata:      m.getCacheMetadata,
		putMetadata:      m.putCacheMetadata,
		changeLeaseOwner: func(*Metadata, uint64) {}, // no-op
	}
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
		case txMetadata := <-m.txnChan:
			logger.Debugf("[%s] Got notification about a Sidetree transaction in block %d and txnNum %d - triggering processing", m.channelID, txMetadata.BlockNum, txMetadata.TxNum)
			m.trigger()
		case <-m.done:
			logger.Infof("[%s] Exiting Observer", m.channelID)
			close(m.processChan)
			return
		}
	}
}

func (m *Observer) trigger() {
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

func (m *Observer) changeLeaseOwner(metadata *Metadata, blockNum uint64) {
	metadata.LeaseOwner = m.leaseProvider.CreateLease(blockNum).Owner()

	logger.Debugf("[%s] Done processing blocks. Lease owner for the next block [%d] is [%s]", m.channelID, blockNum, metadata.LeaseOwner)
}

func (m *Observer) processTxn(sidetreeTxn *txn.SidetreeTxn, pv protocol.Version) error {
	if err := pv.TransactionProcessor().Process(*sidetreeTxn); err != nil {
		return errors.WithMessagef(err, "error processing Txn for anchor [%s] in block [%d] and TxNum [%d]", sidetreeTxn.AnchorString, sidetreeTxn.TransactionTime, sidetreeTxn.TransactionNumber)
	}

	return nil
}

func (m *Observer) processTxnForCache(sidetreeTxn *txn.SidetreeTxn, pv protocol.Version) error {
	docCache, err := m.cacheInvalidatorProvider.GetDocumentInvalidator(m.channelID, sidetreeTxn.Namespace)
	if err != nil {
		logger.Errorf("[%s:%s] Error retrieving document cache invalidator: %s", m.channelID, sidetreeTxn.Namespace, err)

		// Don't return an error since invalidating the cache is not fatal
		return nil
	}

	ops, err := pv.OperationProvider().GetTxnOperations(sidetreeTxn)
	if err != nil {
		logger.Errorf("[%s:%s] Got error retrieving operations for anchor [%s] for the purpose of updating the document cache: %s", m.channelID, sidetreeTxn.Namespace, sidetreeTxn.AnchorString, sidetreeTxn.TransactionTime, sidetreeTxn.TransactionNumber, err)

		// Don't return an error since invalidating the cache is not fatal
		return nil
	}

	for _, op := range ops {
		logger.Debugf("[%s:%s] Invalidating cached document [%s]", m.channelID, sidetreeTxn.Namespace, op.UniqueSuffix)

		docCache.Invalidate(op.UniqueSuffix)
	}

	return nil
}

func (m *Observer) getMetadata() (Metadata, bool, error) {
	metadata, err := m.metadataStore.Get()
	if err != nil {
		if err == errMetaDataNotFound {
			metadata = newMetadata(m.leaseProvider.CreateLease(1).Owner(), 0)
		} else {
			return Metadata{}, false, err
		}
	}

	return *metadata, m.checkLeaseOwner(metadata), nil
}

func (m *Observer) getCacheMetadata() (Metadata, bool, error) {
	var cacheMetadata Metadata

	m.mutex.RLock()
	cacheMetadata = m.cacheMetadata
	m.mutex.RUnlock()

	if cacheMetadata.LastBlockProcessed == 0 {
		logger.Debugf("[%s] Document cache metadata not found. Loading from database", m.channelID)

		metadata, err := m.metadataStore.Get()
		if err != nil {
			if err == errMetaDataNotFound {
				logger.Debugf("[%s] Metadata not found in database. Creating new cache metadata", m.channelID)

				metadata = newMetadata(m.leaseProvider.CreateLease(1).Owner(), 0)
			} else {
				logger.Errorf("[%s] Error retrieving metadata: %s", m.channelID, err)

				// Don't return an error since invalidating the cache is not fatal
				return cacheMetadata, false, nil
			}
		}

		m.putCacheMetadata(metadata)

		cacheMetadata = *metadata
	}

	return cacheMetadata, true, nil
}

func (m *Observer) putMetadata(metadata *Metadata) {
	logger.Debugf("[%s] Updating metadata: %+v", m.channelID, metadata)

	if err := m.metadataStore.Put(metadata); err != nil {
		logger.Errorf("[%s] Error saving metadata: %s", m.channelID, err)
	}
}

func (m *Observer) putCacheMetadata(metadata *Metadata) {
	logger.Debugf("[%s] Updating document cache metadata %+v", m.channelID, metadata)

	m.mutex.Lock()
	m.cacheMetadata = *metadata
	m.mutex.Unlock()
}

var isObserver = func() bool { return role.IsObserver() }
var isResolver = func() bool { return role.IsResolver() }
