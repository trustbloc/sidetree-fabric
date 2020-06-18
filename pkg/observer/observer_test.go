/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package observer

import (
	"encoding/json"
	"fmt"
	"github.com/trustbloc/sidetree-core-go/pkg/commitment"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"

	cb "github.com/hyperledger/fabric-protos-go/common"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	gossipapi "github.com/hyperledger/fabric/extensions/gossip/api"
	gcommon "github.com/hyperledger/fabric/gossip/common"
	peerextmocks "github.com/trustbloc/fabric-peer-ext/pkg/mocks"
	"github.com/trustbloc/fabric-peer-ext/pkg/roles"
	"github.com/trustbloc/sidetree-core-go/pkg/api/batch"
	"github.com/trustbloc/sidetree-core-go/pkg/api/protocol"
	"github.com/trustbloc/sidetree-core-go/pkg/compression"
	"github.com/trustbloc/sidetree-core-go/pkg/docutil"
	"github.com/trustbloc/sidetree-core-go/pkg/jws"
	coremocks "github.com/trustbloc/sidetree-core-go/pkg/mocks"
	"github.com/trustbloc/sidetree-core-go/pkg/operation"
	"github.com/trustbloc/sidetree-core-go/pkg/restapi/helper"
	"github.com/trustbloc/sidetree-core-go/pkg/txnhandler"
	"github.com/trustbloc/sidetree-core-go/pkg/txnhandler/models"

	"github.com/trustbloc/sidetree-fabric/pkg/config"
	ctxcommon "github.com/trustbloc/sidetree-fabric/pkg/context/common"
	stmocks "github.com/trustbloc/sidetree-fabric/pkg/mocks"
	"github.com/trustbloc/sidetree-fabric/pkg/observer/common"
	obmocks "github.com/trustbloc/sidetree-fabric/pkg/observer/mocks"
	"github.com/trustbloc/sidetree-fabric/pkg/peer/mocks"
	"github.com/trustbloc/sidetree-fabric/pkg/role"
)

const (
	channel1 = "channel1"
	org1     = "Org1MSP"

	peer1 = "peer1.org1.com"
	peer2 = "peer2.org1.com"
	peer3 = "peer3.org1.com"

	txID1          = "tx1"
	anchor1        = "anchor1"
	namespace      = "did:sidetree"
	monitorPeriod  = 50 * time.Millisecond
	sleepTime      = 200 * time.Millisecond
	metaDataCCName = "document"

	sha2_256 = 18
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
	opStoreProvider := &stmocks.OperationStoreProvider{}
	opStoreProvider.ForNamespaceReturns(&stmocks.OperationStore{}, nil)

	meta := newMetadata(peer1, 1001)
	metaBytes, err := json.Marshal(meta)
	require.NoError(t, err)

	t.Run("Triggered by block event", func(t *testing.T) {
		clients := newMockClients(t)

		require.NoError(t, clients.offLedger.Put(metaDataCCName, MetaDataColName, peer1, metaBytes))

		cfg := config.Observer{
			Period:                10 * time.Second,
			MetaDataChaincodeName: metaDataCCName,
		}

		txnChan := make(chan gossipapi.TxMetadata, 1)
		m := newObserverWithMocks(t, channel1, cfg, clients, opStoreProvider, txnChan)

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
		clients := newMockClients(t)

		require.NoError(t, clients.offLedger.Put(metaDataCCName, MetaDataColName, peer1, metaBytes))

		cfg := config.Observer{
			Period:                monitorPeriod,
			MetaDataChaincodeName: metaDataCCName,
		}

		txnChan := make(chan gossipapi.TxMetadata, 1)
		m := newObserverWithMocks(t, channel1, cfg, clients, opStoreProvider, txnChan)

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
		m := newObserverWithMocks(t, channel1, cfg, clients, opStoreProvider, txnChan)

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

		roles.SetRoles(map[roles.Role]struct{}{roles.CommitterRole: {}, role.Observer: {}})
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
		m := newObserverWithMocks(t, channel1, cfg, clients, opStoreProvider, txnChan)

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
}

func TestObserver_Error(t *testing.T) {
	cfg := config.Observer{
		Period:                monitorPeriod,
		MetaDataChaincodeName: metaDataCCName,
	}

	ad := txnhandler.AnchorData{
		AnchorAddress:      anchor1,
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

	t.Run("Blockchain.ForChannel error", func(t *testing.T) {
		clients := newMockClients(t)
		clients.blockchainProvider.ForChannelReturns(nil, errors.New("blockchain.ForChannel error"))

		m := newObserverWithMocks(t, channel1, cfg, clients, opStoreProvider, txnChan)
		require.NoError(t, m.Start())
		time.Sleep(sleepTime)
		m.Stop()
	})

	t.Run("Blockchain.GetBlockchainInfo error", func(t *testing.T) {
		clients := newMockClients(t)
		clients.blockchain.GetBlockchainInfoReturns(nil, errors.New("blockchain.GetBlockchainInfo error"))

		m := newObserverWithMocks(t, channel1, cfg, clients, opStoreProvider, txnChan)
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
		clients.dcas.GetReturns(nil, errors.New("injected DCAS error"))

		m := newObserverWithMocks(t, channel1, cfg, clients, opStoreProvider, txnChan)
		require.NoError(t, m.Start())
		time.Sleep(sleepTime)
		m.Stop()
	})

	t.Run("off-ledger client error", func(t *testing.T) {
		clients := newMockClients(t)
		clients.offLedger.GetErr = errors.New("injected off-ledger error")

		m := newObserverWithMocks(t, channel1, cfg, clients, opStoreProvider, txnChan)
		require.NoError(t, m.Start())
		time.Sleep(sleepTime)
		m.Stop()
	})

	t.Run("Blockchain.GetBlockByNum error", func(t *testing.T) {
		clients := newMockClients(t)
		clients.blockchain.GetBlockByNumberReturns(nil, errors.New("blockchain.GetBlockByNumber error"))
		bcInfo := &cb.BlockchainInfo{Height: 1002}
		clients.blockchain.GetBlockchainInfoReturns(bcInfo, nil)

		m := newObserverWithMocks(t, channel1, cfg, clients, opStoreProvider, txnChan)

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

		m := newObserverWithMocks(t, channel1, cfg, clients, opStoreProvider, txnChan)

		require.NoError(t, m.Start())
		time.Sleep(sleepTime)
		m.Stop()
	})

	t.Run("getAnchorFile error", func(t *testing.T) {
		clients := newMockClients(t)
		bcInfo := &cb.BlockchainInfo{Height: 1002}
		clients.blockchain.GetBlockchainInfoReturns(bcInfo, nil)
		clients.blockchain.GetBlockByNumberReturns(b.Build(), nil)
		clients.dcas.GetReturnsOnCall(0, nil, errors.New("getAnchorFile error"))

		m := newObserverWithMocks(t, channel1, cfg, clients, opStoreProvider, txnChan)

		require.NoError(t, m.Start())
		time.Sleep(sleepTime)
		m.Stop()
	})

	t.Run("anchor file unmarshal error", func(t *testing.T) {
		clients := newMockClients(t)
		bcInfo := &cb.BlockchainInfo{Height: 1002}
		clients.blockchain.GetBlockchainInfoReturns(bcInfo, nil)
		clients.blockchain.GetBlockByNumberReturns(b.Build(), nil)
		clients.dcas.GetReturns([]byte("invalid anchor file"), nil)

		m := newObserverWithMocks(t, channel1, cfg, clients, opStoreProvider, txnChan)

		require.NoError(t, m.Start())
		time.Sleep(sleepTime)
		m.Stop()
	})

	t.Run("getBatchFile error", func(t *testing.T) {
		clients := newMockClients(t)
		bcInfo := &cb.BlockchainInfo{Height: 1002}
		clients.blockchain.GetBlockchainInfoReturns(bcInfo, nil)
		clients.blockchain.GetBlockByNumberReturns(b.Build(), nil)

		anchorFile := &models.AnchorFile{}
		anchorFileBytes, err := json.Marshal(anchorFile)
		require.NoError(t, err)
		clients.dcas.GetReturnsOnCall(0, anchorFileBytes, nil)
		clients.dcas.GetReturnsOnCall(1, nil, errors.New("get batch file error"))

		m := newObserverWithMocks(t, channel1, cfg, clients, opStoreProvider, txnChan)

		require.NoError(t, m.Start())
		time.Sleep(sleepTime)
		m.Stop()
	})

	t.Run("batch file unmarshal error", func(t *testing.T) {
		clients := newMockClients(t)
		bcInfo := &cb.BlockchainInfo{Height: 1002}
		clients.blockchain.GetBlockchainInfoReturns(bcInfo, nil)
		clients.blockchain.GetBlockByNumberReturns(b.Build(), nil)

		anchorFile := &models.AnchorFile{}
		anchorFileBytes, err := json.Marshal(anchorFile)
		require.NoError(t, err)
		clients.dcas.GetReturnsOnCall(0, anchorFileBytes, nil)
		clients.dcas.GetReturnsOnCall(1, []byte("invalid batch file"), nil)

		m := newObserverWithMocks(t, channel1, cfg, clients, opStoreProvider, txnChan)

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

		anchorFile := &models.AnchorFile{}
		anchorFileBytes, err := json.Marshal(anchorFile)
		require.NoError(t, err)
		clients.dcas.GetReturnsOnCall(0, anchorFileBytes, nil)
		clients.dcas.GetReturnsOnCall(1, []byte("invalid batch file"), nil)

		m := newObserverWithMocks(t, channel1, cfg, clients, opStoreProvider, txnChan)

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

		m := newObserverWithMocks(t, channel1, cfg, clients, opStoreProvider, txnChan)

		require.NoError(t, m.Start())
		time.Sleep(sleepTime)
		m.Stop()
	})

	t.Run("Retry on transient error", func(t *testing.T) {
		clients := newMockClients(t)
		bcInfo := &cb.BlockchainInfo{Height: 1002}
		clients.blockchain.GetBlockchainInfoReturns(bcInfo, nil)
		clients.blockchain.GetBlockByNumberReturns(b.Build(), nil)

		ops, err := getTestOperations(2)
		require.NoError(t, err)

		op1Bytes, err := json.Marshal(ops[0])
		require.NoError(t, err)
		op2Bytes, err := json.Marshal(ops[1])
		require.NoError(t, err)

		anchorFileBytes, err := json.Marshal(models.CreateAnchorFile("", ops))
		require.NoError(t, err)

		clients.dcas.GetReturnsOnCall(0, anchorFileBytes, nil)
		clients.dcas.GetReturnsOnCall(1, nil, errors.New("get batch file error"))
		clients.dcas.GetReturnsOnCall(2, anchorFileBytes, nil)
		clients.dcas.GetReturnsOnCall(3, op1Bytes, nil)
		clients.dcas.GetReturnsOnCall(4, op2Bytes, nil)

		m := newObserverWithMocks(t, channel1, cfg, clients, opStoreProvider, txnChan)

		require.NoError(t, m.Start())
		time.Sleep(sleepTime)
		m.Stop()
	})
}

type mockClients struct {
	offLedgerProvider  *obmocks.OffLedgerClientProvider
	dcasProvider       *stmocks.DCASClientProvider
	blockchainProvider *obmocks.BlockchainClientProvider
	blockchain         *obmocks.BlockchainClient
	offLedger          *obmocks.MockOffLedgerClient
	dcas               *stmocks.DCASClient
}

func newMockClients(t *testing.T) *mockClients {
	clients := &mockClients{}

	clients.offLedgerProvider = &obmocks.OffLedgerClientProvider{}
	clients.dcasProvider = &stmocks.DCASClientProvider{}
	clients.blockchainProvider = &obmocks.BlockchainClientProvider{}

	clients.blockchain = &obmocks.BlockchainClient{}
	clients.blockchainProvider.ForChannelReturns(clients.blockchain, nil)

	clients.offLedger = obmocks.NewMockOffLedgerClient()
	clients.offLedgerProvider.ForChannelReturns(clients.offLedger, nil)

	clients.dcas = &stmocks.DCASClient{}
	clients.dcasProvider.ForChannelReturns(clients.dcas, nil)

	bcInfo := &cb.BlockchainInfo{
		Height: 1003,
	}

	const numOfOps = 2
	ad := &txnhandler.AnchorData{
		AnchorAddress:      anchor1,
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

	ops, err := getTestOperations(numOfOps)
	require.NoError(t, err)

	op1Bytes, err := json.Marshal(ops[0])
	require.NoError(t, err)
	op2Bytes, err := json.Marshal(ops[1])
	require.NoError(t, err)

	anchorFileBytes, err := compress(models.CreateAnchorFile("map", ops))
	require.NoError(t, err)

	mapFileBytes, err := compress(models.CreateMapFile([]string{"chunk"}, ops))
	require.NoError(t, err)

	chunkFileBytes, err := compress(models.CreateChunkFile(ops))
	require.NoError(t, err)

	clients.blockchain.GetBlockchainInfoReturns(bcInfo, nil)
	clients.blockchain.GetBlockByNumberReturns(b.Build(), nil)
	clients.dcas.GetReturnsOnCall(0, anchorFileBytes, nil)
	clients.dcas.GetReturnsOnCall(1, mapFileBytes, nil)
	clients.dcas.GetReturnsOnCall(2, chunkFileBytes, nil)
	clients.dcas.GetReturnsOnCall(3, op1Bytes, nil)
	clients.dcas.GetReturnsOnCall(4, op2Bytes, nil)

	return clients
}

func compress(model interface{}) ([]byte, error) {
	bytes, err := docutil.MarshalCanonical(model)
	if err != nil {
		return nil, err
	}

	cp := compression.New(compression.WithDefaultAlgorithms())

	return cp.Compress("GZIP", bytes)
}

func newObserverWithMocks(t *testing.T, channelID string, cfg config.Observer, clients *mockClients, opStoreProvider ctxcommon.OperationStoreProvider, txnChan <-chan gossipapi.TxMetadata) *Observer {
	peerCfg := &mocks.PeerConfig{}
	peerCfg.MSPIDReturns(org1)
	peerCfg.PeerIDReturns(peer1)

	dcasCfg := config.DCAS{
		ChaincodeName: sideTreeTxnCCName,
		Collection:    dcasColl,
	}

	gossip := peerextmocks.NewMockGossipAdapter()
	gossip.Self(org1, p1Org1).Member(org1, p2Org1)

	gossipProvider := &peerextmocks.GossipProvider{}
	gossipProvider.GetGossipServiceReturns(gossip)

	m := New(
		channelID, peerCfg, cfg, dcasCfg,
		&ClientProviders{
			OffLedger:  clients.offLedgerProvider,
			DCAS:       clients.dcasProvider,
			Blockchain: clients.blockchainProvider,
			Gossip:     gossipProvider,
		},
		opStoreProvider, txnChan, coremocks.NewMockProtocolClientProvider(),
	)
	require.NotNil(t, m)

	return m
}

func getTestOperations(createOpsNum int) ([]*batch.Operation, error) {
	var ops []*batch.Operation
	for j := 1; j <= createOpsNum; j++ {
		op, err := generateCreateOperations(j)
		if err != nil {
			return nil, err
		}

		ops = append(ops, op)
	}

	return ops, nil
}

func generateCreateOperations(num int) (*batch.Operation, error) {
	testKey := &jws.JWK{
		Crv: "crv",
		Kty: "kty",
		X:   "x",
	}

	c, err := commitment.Calculate(testKey, sha2_256)
	if err != nil {
		return nil, err
	}

	doc := fmt.Sprintf(`{"test":%d}`, num)
	info := &helper.CreateRequestInfo{OpaqueDocument: doc,
		RecoveryCommitment: c,
		UpdateCommitment:   c,
		MultihashCode:      sha2_256}

	request, err := helper.NewCreateRequest(info)
	if err != nil {
		return nil, err
	}

	return operation.ParseOperation(namespace, request, protocol.Protocol{
		HashAlgorithmInMultiHashCode: sha2_256,
	})
}
