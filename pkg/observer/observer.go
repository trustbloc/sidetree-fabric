/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package observer

import (
	"strings"
	"sync/atomic"
	"time"

	gossipapi "github.com/hyperledger/fabric/extensions/gossip/api"
	"github.com/pkg/errors"
	"github.com/trustbloc/sidetree-fabric/pkg/common/transienterr"

	cb "github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric/common/flogging"
	"github.com/trustbloc/fabric-peer-ext/pkg/common/blockvisitor"
	"github.com/trustbloc/sidetree-core-go/pkg/observer"

	"github.com/trustbloc/sidetree-fabric/pkg/client"
	"github.com/trustbloc/sidetree-fabric/pkg/config"
	ctxcommon "github.com/trustbloc/sidetree-fabric/pkg/context/common"
	"github.com/trustbloc/sidetree-fabric/pkg/observer/common"
	"github.com/trustbloc/sidetree-fabric/pkg/observer/operationfilter"
)

var logger = flogging.MustGetLogger("sidetree_observer")

const defaultMonitorPeriod = 10 * time.Second

// MetaData contains meta-data for the document store
type MetaData struct {
	LastBlockProcessed uint64
}

// ClientProviders contains the providers for off-ledger, DCAS, and blockchain clients
type ClientProviders struct {
	OffLedger  common.OffLedgerClientProvider
	DCAS       common.DCASClientProvider
	Blockchain common.BlockchainClientProvider
}

type metaDataStore interface {
	Get() (*MetaData, error)
	Put(*MetaData) error
}

// Observer reads blocks from the ledger looking for Sidetree anchor writes and persists the document operations to the document store.
type Observer struct {
	channelID      string
	period         time.Duration
	blockVisitor   *blockvisitor.Visitor
	done           chan struct{}
	txnProcessor   *observer.TxnProcessor
	blockchain     common.BlockchainClientProvider
	metaDataStore  metaDataStore
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
		metaDataStore: NewMetaDataStore(channelID, peerCfg, observerCfg.MetaDataChaincodeName, clientProviders.OffLedger),
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

	lastBlockNum, err := m.lastBlockProcessed()
	if err != nil {
		logger.Warnf("[%s] Error getting last block processed: %s", m.channelID, err)
		return
	}

	logger.Debugf("[%s] Processing - Block height [%d], last block processed [%d]", m.channelID, bcInfo.Height, lastBlockNum)

	for bNum := lastBlockNum + 1; bNum < bcInfo.Height; bNum++ {
		logger.Debugf("[%s] Processing block [%d]", m.channelID, bNum)
		if err = m.processBlock(bNum); err != nil {
			logger.Errorf("[%s] Error processing block [%d]: %s", m.channelID, bNum, err)
			return
		}
	}
}

func (m *Observer) processBlock(bNum uint64) error {
	var block *cb.Block
	var err error

	if block, err = m.getBlockByNumber(bNum); err != nil {
		return errors.WithMessagef(err, "error getting block [%d]", bNum)
	}

	if err = m.blockVisitor.Visit(block); err != nil {
		return err
	}

	if err = m.setLastBlockProcessed(bNum); err != nil {
		return errors.WithMessagef(err, "error setting last block processed for block [%d]", bNum)
	}

	return nil
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

func (m *Observer) lastBlockProcessed() (uint64, error) {
	metaData, err := m.metaDataStore.Get()
	if err != nil {
		if err == errMetaDataNotFound {
			return 0, nil
		}

		return 0, err
	}

	return metaData.LastBlockProcessed, nil
}

func (m *Observer) setLastBlockProcessed(bNum uint64) error {
	metaData := &MetaData{LastBlockProcessed: bNum}

	logger.Debugf("[%s] Updating meta-data: %+v", m.channelID, metaData)

	if err := m.metaDataStore.Put(metaData); err != nil {
		return errors.WithMessage(err, "error persisting meta-data")
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
