/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package common

import (
	gossipapi "github.com/hyperledger/fabric/extensions/gossip/api"
	"github.com/trustbloc/fabric-peer-ext/pkg/collections/client"
	dcasclient "github.com/trustbloc/fabric-peer-ext/pkg/collections/offledger/dcas/client"
	"github.com/trustbloc/sidetree-core-go/pkg/processor"
	bcclient "github.com/trustbloc/sidetree-fabric/pkg/client"
)

//go:generate counterfeiter -o ./../../mocks/dcasclient.gen.go --fake-name DCASClient github.com/trustbloc/fabric-peer-ext/pkg/collections/offledger/dcas/client.DCAS
//go:generate counterfeiter -o ./../../mocks/dcasprovider.gen.go --fake-name DCASClientProvider . DCASClientProvider
//go:generate counterfeiter -o ./../mocks/offledgerprovider.gen.go --fake-name OffLedgerClientProvider . OffLedgerClientProvider
//go:generate counterfeiter -o ./../mocks/bcclientprovider.gen.go --fake-name BlockchainClientProvider . BlockchainClientProvider
//go:generate counterfeiter -o ./../mocks/bcclient.gen.go --fake-name BlockchainClient ../../client Blockchain
//go:generate counterfeiter -o ./../mocks/opstoreclient.gen.go --fake-name OperationStoreClient github.com/trustbloc/sidetree-core-go/pkg/processor.OperationStoreClient
//go:generate counterfeiter -o ./../mocks/opstoreclientprovider.gen.go --fake-name OperationStoreClientProvider . OperationStoreClientProvider

// DCASClientProvider is a DCAS client provider
type DCASClientProvider interface {
	ForChannel(channelID string) (dcasclient.DCAS, error)
}

// OffLedgerClientProvider is an off-ledger client provider
type OffLedgerClientProvider interface {
	ForChannel(channelID string) (client.OffLedger, error)
}

// BlockPublisherProvider provides a block publisher for a given channel
type BlockPublisherProvider interface {
	ForChannel(channelID string) gossipapi.BlockPublisher
}

// BlockchainClientProvider provides a blockchain client for a given channel
type BlockchainClientProvider interface {
	ForChannel(channelID string) (bcclient.Blockchain, error)
}

// OperationStoreClientProvider returns the operation store client for the given channel/namespace
type OperationStoreClientProvider interface {
	Get(channelID, namespace string) processor.OperationStoreClient
}
