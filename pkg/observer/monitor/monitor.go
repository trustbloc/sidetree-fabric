/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package monitor

import (
	"encoding/json"
	"strings"
	"time"

	cb "github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric/common/flogging"
	"github.com/pkg/errors"

	olclient "github.com/trustbloc/fabric-peer-ext/pkg/collections/client"
	"github.com/trustbloc/fabric-peer-ext/pkg/common/blockvisitor"
	"github.com/trustbloc/sidetree-core-go/pkg/observer"

	"github.com/trustbloc/sidetree-fabric/pkg/client"
	ctxcommon "github.com/trustbloc/sidetree-fabric/pkg/context/common"
	"github.com/trustbloc/sidetree-fabric/pkg/observer/common"
	"github.com/trustbloc/sidetree-fabric/pkg/observer/operationfilter"
)

var logger = flogging.MustGetLogger("sidetree_observer")

const (
	// MetaDataColName is the name of the meta-data collection used by the Monitor
	// to store peer-specific information
	MetaDataColName = "meta_data"
)

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

// Monitor maintains multiple document monitors - one for each channel. A document monitor ensures that the peer
// has all of the document operations that it's supposed to have. The monitor reads blocks from the ledger looking
// for batch writes. It traverses through all of the document operations in the batch and ensures that the peer
// has the operations stored in the DCAS document store.
type Monitor struct {
	*ClientProviders
	channelID    string
	peerID       string
	period       time.Duration
	blockVisitor *blockvisitor.Visitor
	done         chan struct{}
	txnProcessor *observer.TxnProcessor
}

// New returns a new document monitor
func New(channelID, localPeerID string, period time.Duration, clientProviders *ClientProviders, opStoreProvider ctxcommon.OperationStoreProvider) *Monitor {
	m := &Monitor{
		channelID:       channelID,
		peerID:          localPeerID,
		period:          period,
		ClientProviders: clientProviders,
		txnProcessor: observer.NewTxnProcessor(
			&observer.Providers{
				DCASClient:       NewSidetreeDCASReader(channelID, clientProviders.DCAS),
				OpStoreProvider:  NewOperationStoreProvider(channelID, opStoreProvider),
				OpFilterProvider: operationfilter.NewProvider(channelID, opStoreProvider),
			},
		),
		done: make(chan struct{}, 1),
	}

	m.blockVisitor = blockvisitor.New(channelID,
		blockvisitor.WithWriteHandler(m.handleWrite),
		blockvisitor.WithErrorHandler(m.handleError),
	)

	return m
}

// Start starts a document monitor
func (m *Monitor) Start() error {
	logger.Infof("[%s] Starting monitor", m.channelID)

	if m.period == 0 {
		logger.Warningf("Document monitor disabled for channel [%s]", m.channelID)
		return nil
	}

	go m.run()

	return nil
}

// Stop stops the document monitor for the given channel
func (m *Monitor) Stop() {
	logger.Infof("[%s] Stopping monitor", m.channelID)

	m.done <- struct{}{}
}

func (m *Monitor) run() {
	logger.Infof("[%s] Starting document monitor with a period of %s", m.channelID, m.period)

	ticker := time.NewTicker(m.period)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := m.check(); err != nil {
				if errors.Cause(err) == client.ErrNoLedger {
					// This happens before the channel is created. Just log an info since it's not serious.
					logger.Infof("[%s] Unable to check blocks since the channel doesn't exist", m.channelID)
				} else {
					logger.Warnf("[%s] Error checking blocks: %s", m.channelID, err)
				}
			}
		case <-m.done:
			logger.Infof("[%s] Exiting document monitor", m.channelID)
			return
		}
	}
}

func (m *Monitor) check() error {
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

func (m *Monitor) checkBlock(bNum uint64) error {
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

func (m *Monitor) handleWrite(w *blockvisitor.Write) error {
	if w.Namespace != common.SidetreeNs {
		logger.Debugf("[%s] Ignoring write to namespace [%s] in block [%d] and TxNum [%d] since it is not the Sidetree namespace [%s]", m.channelID, w.Namespace, w.BlockNum, w.TxNum, common.SidetreeNs)
		return nil
	}
	if !strings.HasPrefix(w.Write.Key, common.AnchorAddrPrefix) {
		logger.Debugf("[%s] Ignoring write to namespace [%s] in block [%d] and TxNum [%d] since the key doesn't have the anchor address prefix [%s]", m.channelID, common.SidetreeNs, w.BlockNum, w.TxNum, common.AnchorAddrPrefix)
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

func (m *Monitor) handleError(err error, ctx *blockvisitor.Context) error {
	if ctx.Category == blockvisitor.UnmarshalErr {
		logger.Errorf("[%s] Ignoring persistent error: %s. Context: %s", m.channelID, err, ctx)
		return nil
	}

	merr, ok := errors.Cause(err).(monitorError)
	if !ok || !merr.Transient() {
		logger.Errorf("[%s] Ignoring persistent error: %s. Context: %s", m.channelID, err, ctx)
		return nil
	}

	logger.Warnf("[%s] Will retry on transient error [%s] in %s. Context: %s", m.channelID, err, m.period, ctx)
	return err
}

func (m *Monitor) getBlockchainInfo() (*cb.BlockchainInfo, error) {
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

func (m *Monitor) getBlockByNumber(bNum uint64) (*cb.Block, error) {
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

func (m *Monitor) blockchainClient() (client.Blockchain, error) {
	return m.Blockchain.ForChannel(m.channelID)
}

func (m *Monitor) offLedgerClient() (olclient.OffLedger, error) {
	return m.OffLedger.ForChannel(m.channelID)
}

func (m *Monitor) lastBlockProcessed() (uint64, error) {
	olClient, err := m.offLedgerClient()
	if err != nil {
		return 0, err
	}
	data, err := olClient.Get(common.DocNs, MetaDataColName, m.peerID)
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

func (m *Monitor) setLastBlockProcessed(bNum uint64) error {
	metaData := &MetaData{LastBlockProcessed: bNum}
	logger.Debugf("[%s] Updating meta-data: %+v", m.channelID, metaData)

	bytes, err := json.Marshal(metaData)
	if err != nil {
		return errors.WithMessage(err, "error marshalling meta-data")
	}

	olClient, err := m.offLedgerClient()
	if err != nil {
		return err
	}

	err = olClient.Put(common.DocNs, MetaDataColName, m.peerID, bytes)
	if err != nil {
		return errors.WithMessage(err, "error persisting meta-data")
	}
	return nil
}
