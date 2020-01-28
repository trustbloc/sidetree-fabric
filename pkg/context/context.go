/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package context

import (
	"github.com/trustbloc/fabric-peer-ext/pkg/collections/offledger/dcas/client"
	txnapi "github.com/trustbloc/fabric-peer-ext/pkg/txn/api"
	protocolApi "github.com/trustbloc/sidetree-core-go/pkg/api/protocol"
	"github.com/trustbloc/sidetree-core-go/pkg/batch"
	"github.com/trustbloc/sidetree-fabric/pkg/context/blockchain"
	"github.com/trustbloc/sidetree-fabric/pkg/context/cas"
	"github.com/trustbloc/sidetree-fabric/pkg/context/protocol"
)

// SidetreeContext implements 'Fabric' version of Sidetree node context
type SidetreeContext struct {
	channelID        string
	namespace        string
	protocolClient   protocolApi.Client
	casClient        batch.CASClient
	blockchainClient batch.BlockchainClient
}

type txnServiceProvider interface {
	ForChannel(channelID string) (txnapi.Service, error)
}

type dcasClientProvider interface {
	ForChannel(channelID string) (client.DCAS, error)
}

// New creates new Sidetree context
func New(channelID, namespace string, protocolVersions map[string]protocolApi.Protocol, txnProvider txnServiceProvider, dcasProvider dcasClientProvider) *SidetreeContext {
	return &SidetreeContext{
		channelID:        channelID,
		namespace:        namespace,
		protocolClient:   protocol.New(protocolVersions),
		casClient:        cas.New(channelID, dcasProvider),
		blockchainClient: blockchain.New(channelID, txnProvider),
	}
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
func (m *SidetreeContext) CAS() batch.CASClient {
	return m.casClient
}
