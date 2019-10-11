/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package monitor

import (
	"encoding/json"
	"strings"
	"sync"
	"time"

	"github.com/hyperledger/fabric/common/flogging"
	cb "github.com/hyperledger/fabric/protos/common"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"github.com/trustbloc/fabric-peer-ext/pkg/common/blockvisitor"
	"github.com/trustbloc/sidetree-core-go/pkg/observer"
	"github.com/trustbloc/sidetree-fabric/pkg/client"
	"github.com/trustbloc/sidetree-fabric/pkg/observer/common"
)

var logger = flogging.MustGetLogger("observer")

const (
	docsMetaDataColName = "meta_data"
)

// MetaData contains meta-data for the document store
type MetaData struct {
	LastBlockProcessed uint64
}

type blockchainClientProvider interface {
	ForChannel(channelID string) (client.Blockchain, error)
}

type dcasClientProvider interface {
	ForChannel(channelID string) client.DCAS
}

type offLedgerClientProvider interface {
	ForChannel(channelID string) client.OffLedger
}

// ClientProviders contains the providers for off-ledger, DCAS, and blockchain clients
type ClientProviders struct {
	OffLedger  offLedgerClientProvider
	DCAS       dcasClientProvider
	Blockchain blockchainClientProvider
}

// Monitor maintains multiple document monitors - one for each channel. A document monitor ensures that the peer
// has all of the document operations that it's supposed to have. The monitor reads blocks from the ledger looking
// for batch writes. It traverses through all of the document operations in the batch and ensures that the peer
// has the operations stored in the DCAS document store.
type Monitor struct {
	*ClientProviders
	monitorByChannel map[string]*channelMonitor
	lock             sync.RWMutex
}

// New returns a new document monitor
func New(clientProviders *ClientProviders) *Monitor {
	return &Monitor{
		ClientProviders:  clientProviders,
		monitorByChannel: make(map[string]*channelMonitor),
	}
}

// Start starts a document monitor for the given channel
func (m *Monitor) Start(channelID string, period time.Duration) error {
	if period == 0 {
		logger.Warningf("Document monitor disabled for channel [%s]", channelID)
		return nil
	}

	m.lock.Lock()
	defer m.lock.Unlock()

	if _, ok := m.monitorByChannel[channelID]; ok {
		return errors.Errorf("monitor for channel [%s] already started", channelID)
	}

	chm, err := newChannelMonitor(channelID, period, m.ClientProviders)
	if err != nil {
		return errors.Errorf("unable to run monitor for channel [%s]: %s", channelID, err)
	}
	go chm.run()

	m.monitorByChannel[channelID] = chm

	return nil
}

// Stop stops the document monitor for the given channel
func (m *Monitor) Stop(channelID string) {
	m.lock.Lock()
	defer m.lock.Unlock()

	chm, ok := m.monitorByChannel[channelID]
	if !ok {
		logger.Warningf("[%s] Monitor not found", channelID)
		return
	}
	chm.stop()
	delete(m.monitorByChannel, channelID)
}

// StopAll stops all document monitors
func (m *Monitor) StopAll() {
	m.lock.Lock()
	defer m.lock.Unlock()

	for _, chm := range m.monitorByChannel {
		chm.stop()
	}

	m.monitorByChannel = make(map[string]*channelMonitor)
}

type channelMonitor struct {
	*ClientProviders
	channelID    string
	peerID       string
	period       time.Duration
	blockVisitor *blockvisitor.Visitor
	done         chan struct{}
	txnProcessor *observer.TxnProcessor
}

func newChannelMonitor(channelID string, period time.Duration, clientProviders *ClientProviders) (*channelMonitor, error) {
	peerID, err := getLocalPeerID()
	if err != nil {
		return nil, err
	}
	m := &channelMonitor{
		ClientProviders: clientProviders,
		channelID:       channelID,
		peerID:          peerID,
		period:          period,
		txnProcessor: observer.NewTxnProcessor(
			NewSidetreeDCASReader(channelID, clientProviders.DCAS),
			NewOperationStore(channelID, clientProviders.DCAS),
		),
		done: make(chan struct{}),
	}
	m.blockVisitor = blockvisitor.New(channelID, blockvisitor.WithWriteHandler(m.handleWrite))
	return m, nil
}

func (m *channelMonitor) stop() {
	m.done <- struct{}{}
}

func (m *channelMonitor) run() {
	logger.Infof("[%s] Starting document monitor with a period of %s", m.channelID, m.period)

	ticker := time.NewTicker(m.period)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := m.check(); err != nil {
				logger.Warnf("[%s] Error checking blocks: %s", m.channelID, err)
			}
		case <-m.done:
			logger.Infof("[%s] Exiting document monitor", m.channelID)
			return
		}
	}
}

func (m *channelMonitor) check() error {
	bcInfo, err := m.getBlockchainInfo()
	if err != nil {
		return err
	}
	lastBlockNum, err := m.lastBlockProcessed()
	if err != nil {
		return err
	}

	logger.Debugf("[%s] Checking documents - Block height [%d], last block processed [%d]", m.channelID, bcInfo.Height, lastBlockNum)

	for bNum := lastBlockNum + 1; bNum < bcInfo.Height; bNum++ {
		logger.Debugf("[%s] Checking block [%d]", m.channelID, bNum)
		if err = m.checkBlock(bNum); err != nil {
			logger.Errorf("[%s] Error checking block [%d]: %s", m.channelID, bNum, err)
			return err
		}
	}

	return nil
}

func (m *channelMonitor) checkBlock(bNum uint64) error {
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

func (m *channelMonitor) handleWrite(w *blockvisitor.Write) error {
	if w.Namespace != common.SidetreeNs {
		logger.Debugf("[%s] Ignoring w to namespace [%s] in block [%d] and TxNum [%d] since it is not the Sidetree namespace [%s]", m.channelID, w.Namespace, w.BlockNum, w.TxNum, common.SidetreeNs)
		return nil
	}
	if !strings.HasPrefix(w.Write.Key, common.AnchorAddrPrefix) {
		logger.Debugf("[%s] Ignoring w to namespace [%s] in block [%d] and TxNum [%d] since the key doesn't have the anchor address prefix [%s]", m.channelID, common.SidetreeNs, w.BlockNum, w.TxNum, common.AnchorAddrPrefix)
		return nil
	}

	logger.Debugf("[%s] Handling w to anchor [%s] in block [%d] and TxNum [%d]", m.channelID, w.Write.Value, w.BlockNum, w.TxNum)
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

func (m *channelMonitor) getBlockchainInfo() (*cb.BlockchainInfo, error) {
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

func (m *channelMonitor) getBlockByNumber(bNum uint64) (*cb.Block, error) {
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

func (m *channelMonitor) blockchainClient() (client.Blockchain, error) {
	return m.Blockchain.ForChannel(m.channelID)
}

func (m *channelMonitor) offLedgerClient() client.OffLedger {
	return m.OffLedger.ForChannel(m.channelID)
}

func (m *channelMonitor) lastBlockProcessed() (uint64, error) {
	data, err := m.offLedgerClient().Get(common.DocNs, docsMetaDataColName, m.peerID)
	if err != nil {
		return 0, errors.WithMessage(err, "error retrieving meta-data")
	}

	if len(data) == 0 {
		logger.Debugf("[%s] No MetaData exists for peer [%s]", m.channelID, m.peerID)
		return 0, nil
	}

	metaData := &MetaData{}
	err = json.Unmarshal(data, metaData)
	if err != nil {
		return 0, errors.WithMessage(err, "error unmarshalling meta-data")
	}

	return metaData.LastBlockProcessed, nil
}

func (m *channelMonitor) setLastBlockProcessed(bNum uint64) error {
	metaData := &MetaData{LastBlockProcessed: bNum}
	logger.Debugf("[%s] Updating meta-data: %+v", m.channelID, metaData)

	bytes, err := json.Marshal(metaData)
	if err != nil {
		return errors.WithMessage(err, "error marshalling meta-data")
	}

	err = m.offLedgerClient().Put(common.DocNs, docsMetaDataColName, m.peerID, bytes)
	if err != nil {
		return errors.WithMessage(err, "error persisting meta-data")
	}
	return nil
}

func getLocalPeerID() (string, error) {
	peerID := viper.GetString("peer.id")
	if peerID == "" {
		return "", errors.New("peer.id isn't set")
	}
	return peerID, nil
}
