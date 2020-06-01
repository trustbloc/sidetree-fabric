/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package common

import (
	gossipapi "github.com/hyperledger/fabric/extensions/gossip/api"
	dcasclient "github.com/trustbloc/fabric-peer-ext/pkg/collections/offledger/dcas/client"

	"github.com/trustbloc/sidetree-core-go/pkg/api/batch"
	"github.com/trustbloc/sidetree-core-go/pkg/api/protocol"
)

// DCASClientProvider is a DCAS client provider
type DCASClientProvider interface {
	ForChannel(channelID string) (dcasclient.DCAS, error)
}

// OperationStore interface to access operation store
type OperationStore interface {
	Get(suffix string) ([]*batch.Operation, error)
	Put(ops []*batch.Operation) error
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
