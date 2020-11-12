/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package context

import (
	"github.com/hyperledger/fabric/core/ledger"
	olclient "github.com/trustbloc/fabric-peer-ext/pkg/collections/client"
	dcasclient "github.com/trustbloc/fabric-peer-ext/pkg/collections/offledger/dcas/client"
	txnapi "github.com/trustbloc/fabric-peer-ext/pkg/txn/api"
	casApi "github.com/trustbloc/sidetree-core-go/pkg/api/cas"
	protocolApi "github.com/trustbloc/sidetree-core-go/pkg/api/protocol"
	"github.com/trustbloc/sidetree-core-go/pkg/batch"
	"github.com/trustbloc/sidetree-core-go/pkg/batch/cutter"
	"github.com/trustbloc/sidetree-core-go/pkg/dochandler"

	"github.com/trustbloc/sidetree-fabric/pkg/config"
	"github.com/trustbloc/sidetree-fabric/pkg/context/blockchain"
	"github.com/trustbloc/sidetree-fabric/pkg/context/protocol"
	"github.com/trustbloc/sidetree-fabric/pkg/rest/sidetreehandler"
)

// SidetreeContext implements 'Fabric' version of Sidetree node context
type SidetreeContext struct {
	channelID        string
	namespace        string
	protocolClient   protocolApi.Client
	casClient        casApi.Client
	blockchainClient batch.BlockchainClient
	opQueue          cutter.OperationQueue
}

type txnServiceProvider interface {
	ForChannel(channelID string) (txnapi.Service, error)
}

type dcasClientProvider interface {
	GetDCASClient(channelID string, namespace string, coll string) (dcasclient.DCAS, error)
}

type offLedgerClientProvider interface {
	ForChannel(channelID string) (olclient.OffLedger, error)
}

type operationQueueProvider interface {
	Create(channelID string, namespace string) (cutter.OperationQueue, error)
}

type ledgerProvider interface {
	GetLedger(cid string) ledger.PeerLedger
}

type cachingOpProcessorProvider interface {
	CreateCachingOperationProcessor(channelID string, cfg sidetreehandler.Config, resolver dochandler.OperationProcessor) dochandler.OperationProcessor
}

// Providers contains the providers required by the SidetreeContext
type Providers struct {
	TxnProvider                txnServiceProvider
	DCASProvider               dcasClientProvider
	OffLedgerProvider          offLedgerClientProvider
	OperationQueueProvider     operationQueueProvider
	LedgerProvider             ledgerProvider
	OperationProcessorProvider cachingOpProcessorProvider
}

// New creates new Sidetree context
func New(
	channelID, namespace string,
	dcasCfg config.DCAS,
	casClient casApi.Client,
	protocolVersions []protocolApi.Version,
	providers *Providers) (*SidetreeContext, error) {
	opQueue, err := providers.OperationQueueProvider.Create(channelID, namespace)
	if err != nil {
		return nil, err
	}

	return &SidetreeContext{
		channelID:        channelID,
		namespace:        namespace,
		protocolClient:   protocol.New(protocolVersions, providers.LedgerProvider.GetLedger(channelID)),
		casClient:        casClient,
		blockchainClient: blockchain.New(channelID, dcasCfg.ChaincodeName, namespace, providers.TxnProvider),
		opQueue:          opQueue,
	}, nil
}

// Namespace returns the namespace
func (m *SidetreeContext) Namespace() string {
	return m.namespace
}

// Protocol returns protocol client
func (m *SidetreeContext) Protocol() protocolApi.Client {
	return m.protocolClient
}

// Blockchain returns blockchain client
func (m *SidetreeContext) Blockchain() batch.BlockchainClient {
	return m.blockchainClient
}

// CAS returns content addressable storage client
func (m *SidetreeContext) CAS() casApi.Client {
	return m.casClient
}

// OperationQueue returns the queue containing the pending operations
func (m *SidetreeContext) OperationQueue() cutter.OperationQueue {
	return m.opQueue
}
