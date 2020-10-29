/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package common

import (
	gossipapi "github.com/hyperledger/fabric/extensions/gossip/api"
	dcasclient "github.com/trustbloc/fabric-peer-ext/pkg/collections/offledger/dcas/client"

	"github.com/trustbloc/sidetree-core-go/pkg/api/operation"
	"github.com/trustbloc/sidetree-core-go/pkg/api/protocol"
)

//go:generate counterfeiter -o ../../mocks/protocolclient.gen.go --fake-name ProtocolClient github.com/trustbloc/sidetree-core-go/pkg/api/protocol.Client
//go:generate counterfeiter -o ../../mocks/protocolclientprovider.gen.go --fake-name ProtocolClientProvider github.com/trustbloc/sidetree-core-go/pkg/api/protocol.ClientProvider
//go:generate counterfeiter -o ../../mocks/operationprovider.gen.go --fake-name OperationProvider github.com/trustbloc/sidetree-core-go/pkg/api/protocol.OperationProvider

// DCASClientProvider is a DCAS client provider
type DCASClientProvider interface {
	ForChannel(channelID string) (dcasclient.DCAS, error)
}

// OperationStore interface to access operation store
type OperationStore interface {
	Get(suffix string) ([]*operation.AnchoredOperation, error)
	Put(ops []*operation.AnchoredOperation) error
}

// OperationStoreProvider returns an operation store for the given namespace
type OperationStoreProvider interface {
	ForNamespace(namespace string) (OperationStore, error)
}

// ProtocolClientProvider returns a protocol provider for the given namespace
type ProtocolClientProvider interface {
	ForNamespace(namespace string) (protocol.Client, error)
}

// BlockPublisherProvider provides a block publisher for a given channel
type BlockPublisherProvider interface {
	ForChannel(channelID string) gossipapi.BlockPublisher
}
