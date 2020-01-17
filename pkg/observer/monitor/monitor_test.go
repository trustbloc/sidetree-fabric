/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package monitor

import (
	"encoding/base64"
	"encoding/json"
	"testing"
	"time"

	cb "github.com/hyperledger/fabric-protos-go/common"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	peerextmocks "github.com/trustbloc/fabric-peer-ext/pkg/mocks"
	"github.com/trustbloc/sidetree-core-go/pkg/api/batch"
	"github.com/trustbloc/sidetree-core-go/pkg/observer"
	stmocks "github.com/trustbloc/sidetree-fabric/pkg/mocks"
	"github.com/trustbloc/sidetree-fabric/pkg/observer/common"
	"github.com/trustbloc/sidetree-fabric/pkg/observer/mocks"
)

const (
	channel1      = "channel1"
	channel2      = "channel2"
	peer1         = "peer1.org1.com"
	txID1         = "tx1"
	anchor1       = "anchor1"
	monitorPeriod = 50 * time.Millisecond
	sleepTime     = 200 * time.Millisecond
)

func TestMonitor(t *testing.T) {
	m, clients := newMonitorWithMocks(t)

	bcInfo := &cb.BlockchainInfo{
		Height: 1002,
	}
	clients.blockchain.GetBlockchainInfoReturns(bcInfo, nil)

	b := peerextmocks.NewBlockBuilder(channel1, 1001)
	tb1 := b.Transaction(txID1, pb.TxValidationCode_VALID)
	tb1.ChaincodeAction(common.SidetreeNs).
		Write(common.AnchorAddrPrefix+anchor1, []byte(anchor1)).
		Write("non_anchor_key", []byte("some value"))
	tb1.ChaincodeAction("some_other_cc").
		Write("some_key", []byte("some value"))
	clients.blockchain.GetBlockByNumberReturns(b.Build(), nil)

	meta := &MetaData{
		LastBlockProcessed: 1000,
	}
	metaBytes, err := json.Marshal(meta)
	require.NoError(t, err)
	require.NoError(t, clients.offLedger.Put(common.DocNs, docsMetaDataColName, peer1, metaBytes))

	op1 := &batch.Operation{
		ID: "op1",
	}
	op1Bytes, err := json.Marshal(op1)
	require.NoError(t, err)
	b64Op1 := base64.URLEncoding.EncodeToString(op1Bytes)

	op2 := &batch.Operation{
		ID: "op2",
	}
	op2Bytes, err := json.Marshal(op2)
	require.NoError(t, err)
	b64Op2 := base64.URLEncoding.EncodeToString(op2Bytes)

	batchFile := &observer.BatchFile{
		Operations: []string{b64Op1, b64Op2},
	}
	batchFileBytes, err := json.Marshal(batchFile)
	require.NoError(t, err)

	anchorFile := &observer.AnchorFile{}
	anchorFileBytes, err := json.Marshal(anchorFile)
	require.NoError(t, err)

	clients.dcas.GetReturnsOnCall(0, anchorFileBytes, nil)
	clients.dcas.GetReturnsOnCall(1, batchFileBytes, nil)
	clients.dcas.GetReturnsOnCall(2, op1Bytes, nil)
	clients.dcas.GetReturnsOnCall(3, op2Bytes, nil)

	t.Run("No peer ID", func(t *testing.T) {
		restore := m.peerID
		defer func() { m.peerID = restore }()

		m.peerID = ""
		require.Error(t, m.Start(channel1, monitorPeriod))
	})

	t.Run("Success", func(t *testing.T) {
		require.NoError(t, m.Start(channel1, monitorPeriod))
		time.Sleep(sleepTime)
		m.Stop(channel1)
		m.Stop("unknown")

		metaBytes, err := clients.offLedger.Get(common.DocNs, docsMetaDataColName, peer1)
		require.NoError(t, err)

		meta := &MetaData{}
		require.NoError(t, json.Unmarshal(metaBytes, meta))
		require.Equal(t, uint64(1001), meta.LastBlockProcessed)
	})

	t.Run("Start multiple times", func(t *testing.T) {
		require.NoError(t, m.Start(channel1, monitorPeriod))
		require.NoError(t, m.Start(channel2, monitorPeriod))
		require.Error(t, m.Start(channel1, monitorPeriod))
		time.Sleep(sleepTime)
		m.StopAll()
	})

	t.Run("Disabled", func(t *testing.T) {
		require.NoError(t, m.Start(channel1, 0))
		time.Sleep(sleepTime)
		m.Stop(channel1)
	})
}

func TestMonitor_Error(t *testing.T) {
	b := peerextmocks.NewBlockBuilder(channel1, 1001)
	tb1 := b.Transaction(txID1, pb.TxValidationCode_VALID)
	tb1.ChaincodeAction(common.SidetreeNs).
		Write(common.AnchorAddrPrefix+anchor1, []byte(anchor1)).
		Write("non_anchor_key", []byte("some value"))
	tb1.ChaincodeAction("some_other_cc").
		Write("some_key", []byte("some value"))

	t.Run("Blockchain.ForChannel error", func(t *testing.T) {
		m, clients := newMonitorWithMocks(t)
		clients.blockchainProvider.ForChannelReturns(nil, errors.New("blockchain.ForChannel error"))
		require.NoError(t, m.Start(channel1, monitorPeriod))
		time.Sleep(sleepTime)
		m.Stop(channel1)
	})

	t.Run("Blockchain.GetBlockchainInfo error", func(t *testing.T) {
		m, clients := newMonitorWithMocks(t)
		clients.blockchain.GetBlockchainInfoReturns(nil, errors.New("blockchain.GetBlockchainInfo error"))
		require.NoError(t, m.Start(channel1, monitorPeriod))
		time.Sleep(sleepTime)
		m.Stop(channel1)
	})

	t.Run("off-ledger client error", func(t *testing.T) {
		m, clients := newMonitorWithMocks(t)
		clients.offLedger.GetErr = errors.New("injected off-ledger error")
		require.NoError(t, m.Start(channel1, monitorPeriod))
		time.Sleep(sleepTime)
		m.Stop(channel1)
	})

	t.Run("Blockchain.GetBlockByNum error", func(t *testing.T) {
		m, clients := newMonitorWithMocks(t)
		bcInfo := &cb.BlockchainInfo{Height: 1002}
		clients.blockchain.GetBlockchainInfoReturns(bcInfo, nil)
		clients.blockchain.GetBlockByNumberReturns(nil, errors.New("blockchain.GetBlockByNumber error"))
		require.NoError(t, m.Start(channel1, monitorPeriod))
		time.Sleep(sleepTime)
		m.Stop(channel1)
	})

	t.Run("getLastBlockProcessed error", func(t *testing.T) {
		m, clients := newMonitorWithMocks(t)
		bcInfo := &cb.BlockchainInfo{Height: 1002}
		clients.blockchain.GetBlockchainInfoReturns(bcInfo, nil)
		clients.blockchain.GetBlockByNumberReturns(b.Build(), nil)
		clients.offLedger.WithGetErrorForKey(common.DocNs, docsMetaDataColName, peer1, errors.New("getLastBlockProcessed error"))

		require.NoError(t, m.Start(channel1, monitorPeriod))
		time.Sleep(sleepTime)
		m.Stop(channel1)
	})

	t.Run("getAnchorFile error", func(t *testing.T) {
		m, clients := newMonitorWithMocks(t)
		bcInfo := &cb.BlockchainInfo{Height: 1002}
		clients.blockchain.GetBlockchainInfoReturns(bcInfo, nil)
		clients.blockchain.GetBlockByNumberReturns(b.Build(), nil)
		clients.dcas.GetReturnsOnCall(0, nil, errors.New("getAnchorFile error"))

		require.NoError(t, m.Start(channel1, monitorPeriod))
		time.Sleep(sleepTime)
		m.Stop(channel1)
	})

	t.Run("anchor file unmarshal error", func(t *testing.T) {
		m, clients := newMonitorWithMocks(t)
		bcInfo := &cb.BlockchainInfo{Height: 1002}
		clients.blockchain.GetBlockchainInfoReturns(bcInfo, nil)
		clients.blockchain.GetBlockByNumberReturns(b.Build(), nil)
		clients.dcas.GetReturns([]byte("invalid anchor file"), nil)

		require.NoError(t, m.Start(channel1, monitorPeriod))
		time.Sleep(sleepTime)
		m.Stop(channel1)
	})

	t.Run("getBatchFile error", func(t *testing.T) {
		m, clients := newMonitorWithMocks(t)
		bcInfo := &cb.BlockchainInfo{Height: 1002}
		clients.blockchain.GetBlockchainInfoReturns(bcInfo, nil)
		clients.blockchain.GetBlockByNumberReturns(b.Build(), nil)

		anchorFile := &observer.AnchorFile{}
		anchorFileBytes, err := json.Marshal(anchorFile)
		require.NoError(t, err)
		clients.dcas.GetReturnsOnCall(0, anchorFileBytes, nil)
		clients.dcas.GetReturnsOnCall(1, nil, errors.New("get batch file error"))

		require.NoError(t, m.Start(channel1, monitorPeriod))
		time.Sleep(sleepTime)
		m.Stop(channel1)
	})

	t.Run("batch file unmarshal error", func(t *testing.T) {
		m, clients := newMonitorWithMocks(t)
		bcInfo := &cb.BlockchainInfo{Height: 1002}
		clients.blockchain.GetBlockchainInfoReturns(bcInfo, nil)
		clients.blockchain.GetBlockByNumberReturns(b.Build(), nil)

		anchorFile := &observer.AnchorFile{}
		anchorFileBytes, err := json.Marshal(anchorFile)
		require.NoError(t, err)
		clients.dcas.GetReturnsOnCall(0, anchorFileBytes, nil)
		clients.dcas.GetReturnsOnCall(1, []byte("invalid batch file"), nil)

		require.NoError(t, m.Start(channel1, monitorPeriod))
		time.Sleep(sleepTime)
		m.Stop(channel1)
	})

	t.Run("operation base64 error", func(t *testing.T) {
		m, clients := newMonitorWithMocks(t)
		bcInfo := &cb.BlockchainInfo{Height: 1002}
		clients.blockchain.GetBlockchainInfoReturns(bcInfo, nil)
		clients.blockchain.GetBlockByNumberReturns(b.Build(), nil)

		anchorFile := &observer.AnchorFile{}
		anchorFileBytes, err := json.Marshal(anchorFile)
		require.NoError(t, err)
		clients.dcas.GetReturnsOnCall(0, anchorFileBytes, nil)

		batchFile := &observer.BatchFile{
			Operations: []string{"invalid base64 operation"},
		}
		batchFileBytes, err := json.Marshal(batchFile)
		require.NoError(t, err)

		clients.dcas.GetReturnsOnCall(1, batchFileBytes, nil)

		require.NoError(t, m.Start(channel1, monitorPeriod))
		time.Sleep(sleepTime)
		m.Stop(channel1)
	})

	t.Run("operation unmarshal error", func(t *testing.T) {
		m, clients := newMonitorWithMocks(t)
		bcInfo := &cb.BlockchainInfo{Height: 1002}
		clients.blockchain.GetBlockchainInfoReturns(bcInfo, nil)
		clients.blockchain.GetBlockByNumberReturns(b.Build(), nil)

		anchorFile := &observer.AnchorFile{}
		anchorFileBytes, err := json.Marshal(anchorFile)
		require.NoError(t, err)
		clients.dcas.GetReturnsOnCall(0, anchorFileBytes, nil)

		b64Op1 := base64.URLEncoding.EncodeToString([]byte("invalid operation"))

		batchFile := &observer.BatchFile{
			Operations: []string{b64Op1},
		}
		batchFileBytes, err := json.Marshal(batchFile)
		require.NoError(t, err)

		clients.dcas.GetReturnsOnCall(1, batchFileBytes, nil)

		require.NoError(t, m.Start(channel1, monitorPeriod))
		time.Sleep(sleepTime)
		m.Stop(channel1)
	})
}

type mockClients struct {
	offLedgerProvider  *mocks.OffLedgerClientProvider
	dcasProvider       *stmocks.DCASClientProvider
	blockchainProvider *mocks.BlockchainClientProvider
	blockchain         *mocks.BlockchainClient
	offLedger          *mocks.MockOffLedgerClient
	dcas               *stmocks.DCASClient
}

func newMonitorWithMocks(t *testing.T) (*Monitor, *mockClients) {
	clients := &mockClients{}

	clients.offLedgerProvider = &mocks.OffLedgerClientProvider{}
	clients.dcasProvider = &stmocks.DCASClientProvider{}
	clients.blockchainProvider = &mocks.BlockchainClientProvider{}

	clients.blockchain = &mocks.BlockchainClient{}
	clients.blockchainProvider.ForChannelReturns(clients.blockchain, nil)

	clients.offLedger = mocks.NewMockOffLedgerClient()
	clients.offLedgerProvider.ForChannelReturns(clients.offLedger, nil)

	clients.dcas = &stmocks.DCASClient{}
	clients.dcasProvider.ForChannelReturns(clients.dcas, nil)

	m := New(
		peer1,
		&ClientProviders{
			OffLedger:  clients.offLedgerProvider,
			DCAS:       clients.dcasProvider,
			Blockchain: clients.blockchainProvider,
		},
	)
	require.NotNil(t, m)

	return m, clients
}
