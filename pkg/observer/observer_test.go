/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package observer

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	cb "github.com/hyperledger/fabric-protos-go/common"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	gossipapi "github.com/hyperledger/fabric/extensions/gossip/api"
	gcommon "github.com/hyperledger/fabric/gossip/common"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	peerextmocks "github.com/trustbloc/fabric-peer-ext/pkg/mocks"
	"github.com/trustbloc/fabric-peer-ext/pkg/roles"
	"github.com/trustbloc/sidetree-core-go/pkg/api/operation"
	"github.com/trustbloc/sidetree-core-go/pkg/api/protocol"
	coremocks "github.com/trustbloc/sidetree-core-go/pkg/mocks"
	"github.com/trustbloc/sidetree-core-go/pkg/versions/0_1/txnprovider"

	"github.com/trustbloc/sidetree-fabric/pkg/common/transienterr"
	"github.com/trustbloc/sidetree-fabric/pkg/config"
	ctxcommon "github.com/trustbloc/sidetree-fabric/pkg/context/common"
	stmocks "github.com/trustbloc/sidetree-fabric/pkg/mocks"
	"github.com/trustbloc/sidetree-fabric/pkg/observer/common"
	obmocks "github.com/trustbloc/sidetree-fabric/pkg/observer/mocks"
	peermocks "github.com/trustbloc/sidetree-fabric/pkg/peer/mocks"
	"github.com/trustbloc/sidetree-fabric/pkg/role"
)

//go:generate counterfeiter -o ./mocks/doccacheinvalidatorprovider.gen.go --fake-name DocCacheInvalidatorProvider . cacheInvalidatorProvider
//go:generate counterfeiter -o ./mocks/doccacheinvalidator.gen.go --fake-name DocCacheInvalidator ../context/doccache Invalidator

const (
	channel1 = "channel1"
	org1     = "Org1MSP"

	peer1 = "peer1.org1.com"
	peer2 = "peer2.org1.com"
	peer3 = "peer3.org1.com"

	txID1          = "tx1"
	coreIndexURI   = "coreIndexURI"
	namespace      = "did:sidetree"
	monitorPeriod  = 50 * time.Millisecond
	sleepTime      = 200 * time.Millisecond
	metaDataCCName = "document"

	sha2_256 = 18

	sideTreeTxnCCName = "sidetreetxn_cc"
)

var (
	p1Org1PKIID = gcommon.PKIidType("pkiid_P1O1")
	p2Org1PKIID = gcommon.PKIidType("pkiid_P2O1")

	p1Org1 = peerextmocks.NewMember(peer1, p1Org1PKIID)
	p2Org1 = peerextmocks.NewMember(peer2, p2Org1PKIID, role.Observer)

	// Ensure roles are initialized
	_ = roles.GetRoles()
)

func TestObserver(t *testing.T) {
	meta := newMetadata(peer1, 1001)
	metaBytes, err := json.Marshal(meta)
	require.NoError(t, err)

	t.Run("Triggered by block event", func(t *testing.T) {
		restore := setRoles(true, false)
		defer restore()

		clients := newMockClients(t)

		require.NoError(t, clients.offLedger.Put(metaDataCCName, MetaDataColName, peer1, metaBytes))

		cfg := config.Observer{
			Period:                10 * time.Second,
			MetaDataChaincodeName: metaDataCCName,
		}

		txnChan := make(chan gossipapi.TxMetadata, 1)
		m := newObserverWithMocks(t, channel1, cfg, clients, txnChan)

		require.NoError(t, m.Start())
		time.Sleep(sleepTime)

		txnChan <- gossipapi.TxMetadata{BlockNum: 1002, TxNum: 0, ChannelID: channel1, TxID: txID1}
		txnChan <- gossipapi.TxMetadata{BlockNum: 1002, TxNum: 1, ChannelID: channel1, TxID: txID1}
		txnChan <- gossipapi.TxMetadata{BlockNum: 1002, TxNum: 2, ChannelID: channel1, TxID: txID1}
		txnChan <- gossipapi.TxMetadata{BlockNum: 1002, TxNum: 3, ChannelID: channel1, TxID: txID1}

		time.Sleep(sleepTime)
		m.Stop()

		metaBytes, err := clients.offLedger.Get(cfg.MetaDataChaincodeName, MetaDataColName, peer1)
		require.NoError(t, err)

		meta := &Metadata{}
		require.NoError(t, json.Unmarshal(metaBytes, meta))
		require.Equal(t, uint64(1002), meta.LastBlockProcessed)
	})

	t.Run("Triggered by schedule", func(t *testing.T) {
		restore := setRoles(true, false)
		defer restore()

		clients := newMockClients(t)

		require.NoError(t, clients.offLedger.Put(metaDataCCName, MetaDataColName, peer1, metaBytes))

		cfg := config.Observer{
			Period:                monitorPeriod,
			MetaDataChaincodeName: metaDataCCName,
		}

		txnChan := make(chan gossipapi.TxMetadata, 1)
		m := newObserverWithMocks(t, channel1, cfg, clients, txnChan)

		require.NoError(t, m.Start())
		time.Sleep(sleepTime)
		m.Stop()

		metaBytes, err := clients.offLedger.Get(cfg.MetaDataChaincodeName, MetaDataColName, peer1)
		require.NoError(t, err)

		meta := &Metadata{}
		require.NoError(t, json.Unmarshal(metaBytes, meta))
		require.Equal(t, uint64(1002), meta.LastBlockProcessed)
	})

	t.Run("Clustered - replacing lease owner", func(t *testing.T) {
		clients := newMockClients(t)

		roles.SetRoles(map[roles.Role]struct{}{roles.CommitterRole: {}, role.Observer: {}})
		defer roles.SetRoles(nil)

		require.True(t, roles.IsClustered())

		meta := newMetadata(peer3, 1001)

		metaBytes, err := json.Marshal(meta)
		require.NoError(t, err)
		require.NoError(t, clients.offLedger.Put(metaDataCCName, MetaDataColName, org1, metaBytes))

		cfg := config.Observer{
			Period:                monitorPeriod,
			MetaDataChaincodeName: metaDataCCName,
		}

		txnChan := make(chan gossipapi.TxMetadata, 1)
		m := newObserverWithMocks(t, channel1, cfg, clients, txnChan)

		require.NoError(t, m.Start())
		time.Sleep(sleepTime)
		m.Stop()

		metaBytes, err = clients.offLedger.Get(metaDataCCName, MetaDataColName, org1)
		require.NoError(t, err)

		// peer2 should end up the lease owner for the next block
		meta = &Metadata{}
		require.NoError(t, json.Unmarshal(metaBytes, meta))
		require.Equal(t, uint64(1002), meta.LastBlockProcessed)
		require.Equal(t, peer2, meta.LeaseOwner)
	})

	t.Run("Clustered - not lease owner", func(t *testing.T) {
		clients := newMockClients(t)

		roles.SetRoles(map[roles.Role]struct{}{roles.CommitterRole: {}, role.Observer: {}, role.Resolver: {}})
		defer roles.SetRoles(nil)

		require.True(t, roles.IsClustered())

		meta := newMetadata(peer3, 1000)
		metaBytes, err := json.Marshal(meta)
		require.NoError(t, err)
		require.NoError(t, clients.offLedger.Put(metaDataCCName, MetaDataColName, org1, metaBytes))

		cfg := config.Observer{
			Period:                monitorPeriod,
			MetaDataChaincodeName: metaDataCCName,
		}

		txnChan := make(chan gossipapi.TxMetadata, 1)
		m := newObserverWithMocks(t, channel1, cfg, clients, txnChan)

		require.NoError(t, m.Start())
		time.Sleep(sleepTime)
		m.Stop()

		metaBytes, err = clients.offLedger.Get(metaDataCCName, MetaDataColName, org1)
		require.NoError(t, err)

		// The metadata should not have been persisted since the local peer is not the new lease owner
		meta = &Metadata{}
		require.NoError(t, json.Unmarshal(metaBytes, meta))
		require.Equal(t, uint64(1000), meta.LastBlockProcessed)
		require.Equal(t, peer3, meta.LeaseOwner)
	})

	t.Run("DocumentCache invalidator", func(t *testing.T) {
		restore := setRoles(false, true)
		defer restore()

		clients := newMockClients(t)

		require.NoError(t, clients.offLedger.Put(metaDataCCName, MetaDataColName, peer1, metaBytes))

		cfg := config.Observer{
			Period:                10 * time.Second,
			MetaDataChaincodeName: metaDataCCName,
		}

		txnChan := make(chan gossipapi.TxMetadata, 1)
		m := newObserverWithMocks(t, channel1, cfg, clients, txnChan)

		require.NoError(t, m.Start())
		time.Sleep(sleepTime)

		txnChan <- gossipapi.TxMetadata{BlockNum: 1002, TxNum: 0, ChannelID: channel1, TxID: txID1}

		time.Sleep(sleepTime)
		m.Stop()

		meta, _, err := m.getCacheMetadata()
		require.NoError(t, err)

		require.Equal(t, uint64(1002), meta.LastBlockProcessed)
	})
}

func TestObserver_Error(t *testing.T) {
	restore := setRoles(true, true)
	defer restore()

	cfg := config.Observer{
		Period:                monitorPeriod,
		MetaDataChaincodeName: metaDataCCName,
	}

	ad := txnprovider.AnchorData{
		CoreIndexFileURI:   coreIndexURI,
		NumberOfOperations: 1,
	}

	txn := common.TxnInfo{
		AnchorString: ad.GetAnchorString(),
		Namespace:    namespace,
	}

	txnBytes, err := json.Marshal(txn)
	require.NoError(t, err)

	b := peerextmocks.NewBlockBuilder(channel1, 1001)
	tb1 := b.Transaction(txID1, pb.TxValidationCode_VALID)
	tb1.ChaincodeAction(sideTreeTxnCCName).
		Write(common.AnchorPrefix+txn.AnchorString, txnBytes).
		Write("non_anchor_key", []byte("some value"))
	tb1.ChaincodeAction("some_other_cc").
		Write("some_key", []byte("some value"))

	opStoreProvider := &stmocks.OperationStoreProvider{}
	opStoreProvider.ForNamespaceReturns(&stmocks.OperationStore{}, nil)

	txnChan := make(chan gossipapi.TxMetadata)

	meta := newMetadata(peer1, 1001)
	metaBytes, err := json.Marshal(meta)
	require.NoError(t, err)

	t.Run("Blockchain.ForChannel error", func(t *testing.T) {
		clients := newMockClients(t)
		clients.blockchainProvider.ForChannelReturns(nil, errors.New("blockchain.ForChannel error"))

		m := newObserverWithMocks(t, channel1, cfg, clients, txnChan)
		require.NoError(t, m.Start())
		time.Sleep(sleepTime)
		m.Stop()
	})

	t.Run("Blockchain.GetBlockchainInfo error", func(t *testing.T) {
		clients := newMockClients(t)
		clients.blockchain.GetBlockchainInfoReturns(nil, errors.New("blockchain.GetBlockchainInfo error"))

		m := newObserverWithMocks(t, channel1, cfg, clients, txnChan)
		require.NoError(t, m.Start())
		time.Sleep(sleepTime)
		m.Stop()
	})

	t.Run("DCAS client error", func(t *testing.T) {
		meta := newMetadata("", 1000)
		metaBytes, err := json.Marshal(meta)
		require.NoError(t, err)

		bcInfo := &cb.BlockchainInfo{
			Height: 1002,
		}

		clients := newMockClients(t)
		clients.blockchain.GetBlockchainInfoReturns(bcInfo, nil)
		clients.blockchain.GetBlockByNumberReturns(b.Build(), nil)
		require.NoError(t, clients.offLedger.Put(cfg.MetaDataChaincodeName, MetaDataColName, peer1, metaBytes))

		clients.txnProcessor.ProcessReturns(errors.New("injected DCAS error"))

		m := newObserverWithMocks(t, channel1, cfg, clients, txnChan)
		require.NoError(t, m.Start())
		time.Sleep(sleepTime)
		m.Stop()
	})

	t.Run("off-ledger client error", func(t *testing.T) {
		clients := newMockClients(t)
		clients.offLedger.GetErr = errors.New("injected off-ledger error")

		m := newObserverWithMocks(t, channel1, cfg, clients, txnChan)
		require.NoError(t, m.Start())
		time.Sleep(sleepTime)
		m.Stop()
	})

	t.Run("Blockchain.GetBlockByNum error", func(t *testing.T) {
		clients := newMockClients(t)
		clients.blockchain.GetBlockByNumberReturns(nil, errors.New("blockchain.GetBlockByNumber error"))
		bcInfo := &cb.BlockchainInfo{Height: 1002}
		clients.blockchain.GetBlockchainInfoReturns(bcInfo, nil)

		m := newObserverWithMocks(t, channel1, cfg, clients, txnChan)

		require.NoError(t, m.Start())
		time.Sleep(sleepTime)
		m.Stop()
	})

	t.Run("getLastBlockProcessed error", func(t *testing.T) {
		clients := newMockClients(t)
		bcInfo := &cb.BlockchainInfo{Height: 1002}
		clients.blockchain.GetBlockchainInfoReturns(bcInfo, nil)
		clients.blockchain.GetBlockByNumberReturns(b.Build(), nil)
		clients.offLedger.WithGetErrorForKey(metaDataCCName, MetaDataColName, peer1, errors.New("getLastBlockProcessed error"))

		m := newObserverWithMocks(t, channel1, cfg, clients, txnChan)

		require.NoError(t, m.Start())
		time.Sleep(sleepTime)
		m.Stop()
	})

	t.Run("getAnchorFile error", func(t *testing.T) {
		clients := newMockClients(t)
		bcInfo := &cb.BlockchainInfo{Height: 1002}
		clients.blockchain.GetBlockchainInfoReturns(bcInfo, nil)
		clients.blockchain.GetBlockByNumberReturns(b.Build(), nil)
		clients.txnProcessor.ProcessReturnsOnCall(0, errors.New("getAnchorFile error"))

		m := newObserverWithMocks(t, channel1, cfg, clients, txnChan)

		require.NoError(t, m.Start())
		time.Sleep(sleepTime)
		m.Stop()
	})

	t.Run("block unmarshal error", func(t *testing.T) {
		clients := newMockClients(t)
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

		m := newObserverWithMocks(t, channel1, cfg, clients, txnChan)

		require.NoError(t, m.Start())
		time.Sleep(sleepTime)
		m.Stop()
	})

	t.Run("Put metadata error", func(t *testing.T) {
		roles.SetRoles(map[roles.Role]struct{}{roles.CommitterRole: {}, role.Observer: {}})
		defer roles.SetRoles(nil)

		require.True(t, roles.IsClustered())

		clients := newMockClients(t)
		bcInfo := &cb.BlockchainInfo{Height: 1002}
		clients.blockchain.GetBlockchainInfoReturns(bcInfo, nil)
		clients.blockchain.GetBlockByNumberReturns(&cb.Block{
			Header: &cb.BlockHeader{
				Number: 1002,
			},
			Data: &cb.BlockData{},
		}, nil)

		errExpected := errors.New("injected off-ledger client error")
		clients.offLedger.PutErr = errExpected

		m := newObserverWithMocks(t, channel1, cfg, clients, txnChan)

		require.NoError(t, m.Start())
		time.Sleep(sleepTime)
		m.Stop()
	})

	t.Run("Retry on transient error and succeed", func(t *testing.T) {
		clients := newMockClients(t)
		bcInfo := &cb.BlockchainInfo{Height: 1002}
		clients.blockchain.GetBlockchainInfoReturns(bcInfo, nil)
		clients.blockchain.GetBlockByNumberReturns(b.Build(), nil)

		clients.txnProcessor.ProcessReturnsOnCall(1, transienterr.New(errors.New("get batch file error"), transienterr.CodeBlockchain))
		clients.txnProcessor.ProcessReturnsOnCall(2, transienterr.New(errors.New("get batch file error"), transienterr.CodeBlockchain))

		m := newObserverWithMocks(t, channel1, cfg, clients, txnChan)

		require.NoError(t, m.Start())
		time.Sleep(sleepTime)
		m.Stop()
	})

	t.Run("Retry on transient error and give up", func(t *testing.T) {
		clients := newMockClients(t)
		bcInfo := &cb.BlockchainInfo{Height: 1002}
		clients.blockchain.GetBlockchainInfoReturns(bcInfo, nil)
		clients.blockchain.GetBlockByNumberReturns(b.Build(), nil)

		clients.txnProcessor.ProcessReturnsOnCall(1, transienterr.New(errors.New("get batch file error"), transienterr.CodeBlockchain))
		clients.txnProcessor.ProcessReturnsOnCall(2, transienterr.New(errors.New("get batch file error"), transienterr.CodeBlockchain))
		clients.txnProcessor.ProcessReturnsOnCall(3, transienterr.New(errors.New("get batch file error"), transienterr.CodeBlockchain))
		clients.txnProcessor.ProcessReturnsOnCall(4, transienterr.New(errors.New("get batch file error"), transienterr.CodeBlockchain))

		m := newObserverWithMocks(t, channel1, cfg, clients, txnChan)

		require.NoError(t, m.Start())
		time.Sleep(sleepTime)
		m.Stop()
	})

	t.Run("Document cache error", func(t *testing.T) {
		restore := setRoles(false, true)
		defer restore()

		clients := newMockClients(t)
		clients.cacheProvider = &obmocks.DocCacheInvalidatorProvider{}

		require.NoError(t, clients.offLedger.Put(metaDataCCName, MetaDataColName, peer1, metaBytes))

		errExpected := fmt.Errorf("injected document cache error")
		clients.cacheProvider.GetDocumentInvalidatorReturns(nil, errExpected)

		cfg := config.Observer{
			Period:                10 * time.Second,
			MetaDataChaincodeName: metaDataCCName,
		}

		txnChan := make(chan gossipapi.TxMetadata, 1)
		m := newObserverWithMocks(t, channel1, cfg, clients, txnChan)

		require.NoError(t, m.Start())
		time.Sleep(sleepTime)

		txnChan <- gossipapi.TxMetadata{BlockNum: 1002, TxNum: 0, ChannelID: channel1, TxID: txID1}

		time.Sleep(sleepTime)
		m.Stop()

		cacheMeta, _, err := m.getCacheMetadata()
		require.NoError(t, err)
		require.Equal(t, uint64(1002), cacheMeta.LastBlockProcessed)
	})

	t.Run("Get operations error", func(t *testing.T) {
		restore := setRoles(false, true)
		defer restore()

		clients := newMockClients(t)

		require.NoError(t, clients.offLedger.Put(metaDataCCName, MetaDataColName, peer1, metaBytes))

		opp := &stmocks.OperationProvider{}
		opp.GetTxnOperationsReturns(nil, fmt.Errorf("injected operation provider error"))
		clients.pv.OperationProviderReturns(opp)

		cfg := config.Observer{
			Period:                10 * time.Second,
			MetaDataChaincodeName: metaDataCCName,
		}

		txnChan := make(chan gossipapi.TxMetadata, 1)
		m := newObserverWithMocks(t, channel1, cfg, clients, txnChan)

		require.NoError(t, m.Start())
		time.Sleep(sleepTime)

		txnChan <- gossipapi.TxMetadata{BlockNum: 1002, TxNum: 0, ChannelID: channel1, TxID: txID1}

		time.Sleep(sleepTime)
		m.Stop()

		meta, _, err := m.getCacheMetadata()
		require.NoError(t, err)
		require.Equal(t, uint64(1002), meta.LastBlockProcessed)
	})
}

type mockClients struct {
	offLedgerProvider  *obmocks.OffLedgerClientProvider
	blockchainProvider *obmocks.BlockchainClientProvider
	blockchain         *obmocks.BlockchainClient
	offLedger          *obmocks.MockOffLedgerClient
	pcp                ctxcommon.ProtocolClientProvider
	pv                 *coremocks.ProtocolVersion
	txnProcessor       *coremocks.TxnProcessor
	cacheProvider      *obmocks.DocCacheInvalidatorProvider
}

func newMockClients(t *testing.T) *mockClients {
	clients := &mockClients{}

	clients.offLedgerProvider = &obmocks.OffLedgerClientProvider{}
	clients.blockchainProvider = &obmocks.BlockchainClientProvider{}

	clients.blockchain = &obmocks.BlockchainClient{}
	clients.blockchainProvider.ForChannelReturns(clients.blockchain, nil)

	clients.offLedger = obmocks.NewMockOffLedgerClient()
	clients.offLedgerProvider.ForChannelReturns(clients.offLedger, nil)

	bcInfo := &cb.BlockchainInfo{
		Height: 1003,
	}

	const numOfOps = 2
	ad := &txnprovider.AnchorData{
		CoreIndexFileURI:   coreIndexURI,
		NumberOfOperations: numOfOps,
	}

	txn := common.TxnInfo{
		AnchorString: ad.GetAnchorString(),
		Namespace:    namespace,
	}

	txnBytes, err := json.Marshal(txn)
	require.NoError(t, err)

	b := peerextmocks.NewBlockBuilder(channel1, 1002)
	tb1 := b.Transaction(txID1, pb.TxValidationCode_VALID)
	tb1.ChaincodeAction(sideTreeTxnCCName).
		Write(common.AnchorPrefix+txn.AnchorString, txnBytes).
		Write("non_anchor_key", []byte("some value"))
	tb1.ChaincodeAction("some_other_cc").
		Write("some_key", []byte("some value"))

	pv, txnp := newMockProtocolVersion()

	pc := &stmocks.ProtocolClient{}
	pc.CurrentReturns(pv, nil)
	pc.GetReturns(pv, nil)

	pcp := &stmocks.ProtocolClientProvider{}
	pcp.ForNamespaceReturns(pc, nil)
	clients.pcp = pcp
	clients.pv = pv
	clients.txnProcessor = txnp

	clients.blockchain.GetBlockchainInfoReturns(bcInfo, nil)
	clients.blockchain.GetBlockByNumberReturns(b.Build(), nil)

	cacheProvider := &obmocks.DocCacheInvalidatorProvider{}

	cache := &obmocks.DocCacheInvalidator{}
	cacheProvider.GetDocumentInvalidatorReturns(cache, nil)

	clients.cacheProvider = cacheProvider

	return clients
}

func newObserverWithMocks(t *testing.T, channelID string, cfg config.Observer, clients *mockClients, txnChan <-chan gossipapi.TxMetadata) *Observer {
	peerCfg := &peermocks.PeerConfig{}
	peerCfg.MSPIDReturns(org1)
	peerCfg.PeerIDReturns(peer1)

	gossip := peerextmocks.NewMockGossipAdapter()
	gossip.Self(org1, p1Org1).Member(org1, p2Org1)

	gossipProvider := &peerextmocks.GossipProvider{}
	gossipProvider.GetGossipServiceReturns(gossip)

	m := New(
		channelID, peerCfg, cfg,
		&ClientProviders{
			OffLedger:                clients.offLedgerProvider,
			Blockchain:               clients.blockchainProvider,
			Gossip:                   gossipProvider,
			CacheInvalidatorProvider: clients.cacheProvider,
		},
		txnChan, clients.pcp,
	)
	require.NotNil(t, m)

	return m
}

func newMockProtocolVersion() (*coremocks.ProtocolVersion, *coremocks.TxnProcessor) {
	const maxBatchFileSize = 20000
	const maxOperationByteSize = 2000

	p := protocol.Protocol{
		GenesisTime:                 0,
		MultihashAlgorithm:          sha2_256,
		MaxOperationCount:           2,
		MaxOperationSize:            maxOperationByteSize,
		CompressionAlgorithm:        "GZIP",
		MaxChunkFileSize:            maxBatchFileSize,
		MaxProvisionalIndexFileSize: maxBatchFileSize,
		MaxCoreIndexFileSize:        maxBatchFileSize,
		MaxProofFileSize:            maxBatchFileSize,
		SignatureAlgorithms:         []string{"EdDSA", "ES256"},
		KeyAlgorithms:               []string{"Ed25519", "P-256"},
		Patches:                     []string{"add-public-keys", "remove-public-keys", "add-service-endpoints", "remove-service-endpoints", "ietf-json-patch"},
	}

	tp := &coremocks.TxnProcessor{}
	pv := &coremocks.ProtocolVersion{}
	pv.TransactionProcessorReturns(tp)

	pv.ProtocolReturns(p)

	opp := &stmocks.OperationProvider{}

	ops := []*operation.AnchoredOperation{
		{UniqueSuffix: "doc1"},
		{UniqueSuffix: "doc2"},
	}

	opp.GetTxnOperationsReturns(ops, nil)
	pv.OperationProviderReturns(opp)

	return pv, tp
}

func setRoles(observer, resolver bool) func() {
	restoreObserver := isObserver
	restoreResolver := isResolver

	isObserver = func() bool { return observer }
	isResolver = func() bool { return resolver }

	return func() {
		isObserver = restoreObserver
		isResolver = restoreResolver
	}
}
