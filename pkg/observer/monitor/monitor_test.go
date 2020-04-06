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

	"github.com/trustbloc/sidetree-fabric/pkg/config"
	ctxcommon "github.com/trustbloc/sidetree-fabric/pkg/context/common"
	stmocks "github.com/trustbloc/sidetree-fabric/pkg/mocks"
	"github.com/trustbloc/sidetree-fabric/pkg/observer/common"
	"github.com/trustbloc/sidetree-fabric/pkg/observer/mocks"
)

const (
	channel1       = "channel1"
	peer1          = "peer1.org1.com"
	txID1          = "tx1"
	anchor1        = "anchor1"
	monitorPeriod  = 50 * time.Millisecond
	sleepTime      = 200 * time.Millisecond
	metaDataCCName = "document"
)

func TestMonitor(t *testing.T) {
	bcInfo := &cb.BlockchainInfo{
		Height: 1002,
	}

	b := peerextmocks.NewBlockBuilder(channel1, 1001)
	tb1 := b.Transaction(txID1, pb.TxValidationCode_VALID)
	tb1.ChaincodeAction(sideTreeTxnCCName).
		Write(common.AnchorAddrPrefix+anchor1, []byte(anchor1)).
		Write("non_anchor_key", []byte("some value"))
	tb1.ChaincodeAction("some_other_cc").
		Write("some_key", []byte("some value"))

	meta := &MetaData{
		LastBlockProcessed: 1000,
	}
	metaBytes, err := json.Marshal(meta)
	require.NoError(t, err)

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

	clients := newMockClients()
	clients.blockchain.GetBlockchainInfoReturns(bcInfo, nil)
	clients.blockchain.GetBlockByNumberReturns(b.Build(), nil)

	require.NoError(t, clients.offLedger.Put(metaDataCCName, MetaDataColName, peer1, metaBytes))

	clients.dcas.GetReturnsOnCall(0, anchorFileBytes, nil)
	clients.dcas.GetReturnsOnCall(1, batchFileBytes, nil)
	clients.dcas.GetReturnsOnCall(2, op1Bytes, nil)
	clients.dcas.GetReturnsOnCall(3, op2Bytes, nil)

	opStoreProvider := &stmocks.OperationStoreProvider{}

	t.Run("Success", func(t *testing.T) {
		cfg := config.Monitor{
			Period:                monitorPeriod,
			MetaDataChaincodeName: metaDataCCName,
		}

		m := newMonitorWithMocks(t, channel1, cfg, clients, opStoreProvider)

		require.NoError(t, m.Start())
		time.Sleep(sleepTime)
		m.Stop()

		metaBytes, err := clients.offLedger.Get(cfg.MetaDataChaincodeName, MetaDataColName, peer1)
		require.NoError(t, err)

		meta := &MetaData{}
		require.NoError(t, json.Unmarshal(metaBytes, meta))
		require.Equal(t, uint64(1001), meta.LastBlockProcessed)
	})

	t.Run("Disabled", func(t *testing.T) {
		cfg := config.Monitor{
			Period:                0,
			MetaDataChaincodeName: metaDataCCName,
		}

		m := newMonitorWithMocks(t, channel1, cfg, clients, opStoreProvider)
		require.NoError(t, m.Start())
		time.Sleep(sleepTime)
		m.Stop()
	})
}

func TestMonitor_Error(t *testing.T) {
	cfg := config.Monitor{
		Period:                monitorPeriod,
		MetaDataChaincodeName: metaDataCCName,
	}

	b := peerextmocks.NewBlockBuilder(channel1, 1001)
	tb1 := b.Transaction(txID1, pb.TxValidationCode_VALID)
	tb1.ChaincodeAction(sideTreeTxnCCName).
		Write(common.AnchorAddrPrefix+anchor1, []byte(anchor1)).
		Write("non_anchor_key", []byte("some value"))
	tb1.ChaincodeAction("some_other_cc").
		Write("some_key", []byte("some value"))

	opStoreProvider := &stmocks.OperationStoreProvider{}

	t.Run("Blockchain.ForChannel error", func(t *testing.T) {
		clients := newMockClients()
		clients.blockchainProvider.ForChannelReturns(nil, errors.New("blockchain.ForChannel error"))

		m := newMonitorWithMocks(t, channel1, cfg, clients, opStoreProvider)
		require.NoError(t, m.Start())
		time.Sleep(sleepTime)
		m.Stop()
	})

	t.Run("Blockchain.GetBlockchainInfo error", func(t *testing.T) {
		clients := newMockClients()
		clients.blockchain.GetBlockchainInfoReturns(nil, errors.New("blockchain.GetBlockchainInfo error"))

		m := newMonitorWithMocks(t, channel1, cfg, clients, opStoreProvider)
		require.NoError(t, m.Start())
		time.Sleep(sleepTime)
		m.Stop()
	})

	t.Run("DCAS client error", func(t *testing.T) {
		meta := &MetaData{
			LastBlockProcessed: 1000,
		}
		metaBytes, err := json.Marshal(meta)
		require.NoError(t, err)

		bcInfo := &cb.BlockchainInfo{
			Height: 1002,
		}

		clients := newMockClients()
		clients.blockchain.GetBlockchainInfoReturns(bcInfo, nil)
		clients.blockchain.GetBlockByNumberReturns(b.Build(), nil)
		require.NoError(t, clients.offLedger.Put(cfg.MetaDataChaincodeName, MetaDataColName, peer1, metaBytes))
		clients.dcas.GetReturns(nil, errors.New("injected DCAS error"))

		m := newMonitorWithMocks(t, channel1, cfg, clients, opStoreProvider)
		require.NoError(t, m.Start())
		time.Sleep(sleepTime)
		m.Stop()
	})

	t.Run("off-ledger client error", func(t *testing.T) {
		clients := newMockClients()
		clients.offLedger.GetErr = errors.New("injected off-ledger error")

		m := newMonitorWithMocks(t, channel1, cfg, clients, opStoreProvider)
		require.NoError(t, m.Start())
		time.Sleep(sleepTime)
		m.Stop()
	})

	t.Run("Blockchain.GetBlockByNum error", func(t *testing.T) {
		clients := newMockClients()
		clients.blockchain.GetBlockByNumberReturns(nil, errors.New("blockchain.GetBlockByNumber error"))
		bcInfo := &cb.BlockchainInfo{Height: 1002}
		clients.blockchain.GetBlockchainInfoReturns(bcInfo, nil)

		m := newMonitorWithMocks(t, channel1, cfg, clients, opStoreProvider)

		require.NoError(t, m.Start())
		time.Sleep(sleepTime)
		m.Stop()
	})

	t.Run("getLastBlockProcessed error", func(t *testing.T) {
		clients := newMockClients()
		bcInfo := &cb.BlockchainInfo{Height: 1002}
		clients.blockchain.GetBlockchainInfoReturns(bcInfo, nil)
		clients.blockchain.GetBlockByNumberReturns(b.Build(), nil)
		clients.offLedger.WithGetErrorForKey(metaDataCCName, MetaDataColName, peer1, errors.New("getLastBlockProcessed error"))

		m := newMonitorWithMocks(t, channel1, cfg, clients, opStoreProvider)

		require.NoError(t, m.Start())
		time.Sleep(sleepTime)
		m.Stop()
	})

	t.Run("getAnchorFile error", func(t *testing.T) {
		clients := newMockClients()
		bcInfo := &cb.BlockchainInfo{Height: 1002}
		clients.blockchain.GetBlockchainInfoReturns(bcInfo, nil)
		clients.blockchain.GetBlockByNumberReturns(b.Build(), nil)
		clients.dcas.GetReturnsOnCall(0, nil, errors.New("getAnchorFile error"))

		m := newMonitorWithMocks(t, channel1, cfg, clients, opStoreProvider)

		require.NoError(t, m.Start())
		time.Sleep(sleepTime)
		m.Stop()
	})

	t.Run("anchor file unmarshal error", func(t *testing.T) {
		clients := newMockClients()
		bcInfo := &cb.BlockchainInfo{Height: 1002}
		clients.blockchain.GetBlockchainInfoReturns(bcInfo, nil)
		clients.blockchain.GetBlockByNumberReturns(b.Build(), nil)
		clients.dcas.GetReturns([]byte("invalid anchor file"), nil)

		m := newMonitorWithMocks(t, channel1, cfg, clients, opStoreProvider)

		require.NoError(t, m.Start())
		time.Sleep(sleepTime)
		m.Stop()
	})

	t.Run("getBatchFile error", func(t *testing.T) {
		clients := newMockClients()
		bcInfo := &cb.BlockchainInfo{Height: 1002}
		clients.blockchain.GetBlockchainInfoReturns(bcInfo, nil)
		clients.blockchain.GetBlockByNumberReturns(b.Build(), nil)

		anchorFile := &observer.AnchorFile{}
		anchorFileBytes, err := json.Marshal(anchorFile)
		require.NoError(t, err)
		clients.dcas.GetReturnsOnCall(0, anchorFileBytes, nil)
		clients.dcas.GetReturnsOnCall(1, nil, errors.New("get batch file error"))

		m := newMonitorWithMocks(t, channel1, cfg, clients, opStoreProvider)

		require.NoError(t, m.Start())
		time.Sleep(sleepTime)
		m.Stop()
	})

	t.Run("batch file unmarshal error", func(t *testing.T) {
		clients := newMockClients()
		bcInfo := &cb.BlockchainInfo{Height: 1002}
		clients.blockchain.GetBlockchainInfoReturns(bcInfo, nil)
		clients.blockchain.GetBlockByNumberReturns(b.Build(), nil)

		anchorFile := &observer.AnchorFile{}
		anchorFileBytes, err := json.Marshal(anchorFile)
		require.NoError(t, err)
		clients.dcas.GetReturnsOnCall(0, anchorFileBytes, nil)
		clients.dcas.GetReturnsOnCall(1, []byte("invalid batch file"), nil)

		m := newMonitorWithMocks(t, channel1, cfg, clients, opStoreProvider)

		require.NoError(t, m.Start())
		time.Sleep(sleepTime)
		m.Stop()
	})

	t.Run("operation base64 error", func(t *testing.T) {
		clients := newMockClients()
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

		m := newMonitorWithMocks(t, channel1, cfg, clients, opStoreProvider)

		require.NoError(t, m.Start())
		time.Sleep(sleepTime)
		m.Stop()
	})

	t.Run("operation unmarshal error", func(t *testing.T) {
		clients := newMockClients()
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

		m := newMonitorWithMocks(t, channel1, cfg, clients, opStoreProvider)

		require.NoError(t, m.Start())
		time.Sleep(sleepTime)
		m.Stop()
	})

	t.Run("block unmarshal error", func(t *testing.T) {
		clients := newMockClients()
		bcInfo := &cb.BlockchainInfo{Height: 1002}
		clients.blockchain.GetBlockchainInfoReturns(bcInfo, nil)
		clients.blockchain.GetBlockByNumberReturns(&cb.Block{
			Header: &cb.BlockHeader{
				Number: 1002,
			},
			Data: &cb.BlockData{
				Data: [][]byte{[]byte("invalid block data")},
			},
		}, nil)

		anchorFile := &observer.AnchorFile{}
		anchorFileBytes, err := json.Marshal(anchorFile)
		require.NoError(t, err)
		clients.dcas.GetReturnsOnCall(0, anchorFileBytes, nil)
		clients.dcas.GetReturnsOnCall(1, []byte("invalid batch file"), nil)

		m := newMonitorWithMocks(t, channel1, cfg, clients, opStoreProvider)

		require.NoError(t, m.Start())
		time.Sleep(sleepTime)
		m.Stop()
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

func newMockClients() *mockClients {
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

	return clients
}

func newMonitorWithMocks(t *testing.T, channelID string, cfg config.Monitor, clients *mockClients, opStoreProvider ctxcommon.OperationStoreProvider) *Monitor {
	dcasCfg := config.DCAS{
		ChaincodeName: sideTreeTxnCCName,
		Collection:    dcasColl,
	}

	m := New(
		channelID, peer1, cfg, dcasCfg,
		&ClientProviders{
			OffLedger:  clients.offLedgerProvider,
			DCAS:       clients.dcasProvider,
			Blockchain: clients.blockchainProvider,
		},
		opStoreProvider,
	)
	require.NotNil(t, m)

	return m
}
