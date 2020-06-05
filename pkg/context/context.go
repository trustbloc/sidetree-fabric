/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package context

import (
	"github.com/trustbloc/fabric-peer-ext/pkg/collections/offledger/dcas/client"
	txnapi "github.com/trustbloc/fabric-peer-ext/pkg/txn/api"

	casApi "github.com/trustbloc/sidetree-core-go/pkg/api/cas"
	protocolApi "github.com/trustbloc/sidetree-core-go/pkg/api/protocol"
	"github.com/trustbloc/sidetree-core-go/pkg/batch"
	"github.com/trustbloc/sidetree-core-go/pkg/batch/cutter"

	"github.com/trustbloc/sidetree-fabric/pkg/config"
	"github.com/trustbloc/sidetree-fabric/pkg/context/blockchain"
	"github.com/trustbloc/sidetree-fabric/pkg/context/cas"
	"github.com/trustbloc/sidetree-fabric/pkg/context/protocol"
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
	ForChannel(channelID string) (client.DCAS, error)
}

type operationQueueProvider interface {
	Create(channelID string, namespace string) (cutter.OperationQueue, error)
}

// New creates new Sidetree context
func New(
	channelID, namespace string,
	dcasCfg config.DCAS,
	protocolVersions map[string]protocolApi.Protocol,
	txnProvider txnServiceProvider,
	dcasProvider dcasClientProvider,
	opQueueProvider operationQueueProvider) (*SidetreeContext, error) {
	opQueue, err := opQueueProvider.Create(channelID, namespace)
	if err != nil {
		return nil, err
	}

	return &SidetreeContext{
		channelID:        channelID,
		namespace:        namespace,
		protocolClient:   protocol.New(protocolVersions),
		casClient:        cas.New(channelID, dcasCfg, dcasProvider),
		blockchainClient: blockchain.New(channelID, dcasCfg.ChaincodeName, namespace, txnProvider),
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
