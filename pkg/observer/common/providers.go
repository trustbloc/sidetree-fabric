/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package common

import (
	"github.com/trustbloc/fabric-peer-ext/pkg/collections/client"
	dcasclient "github.com/trustbloc/fabric-peer-ext/pkg/collections/offledger/dcas/client"
	bcclient "github.com/trustbloc/sidetree-fabric/pkg/client"
)

//go:generate counterfeiter -o ./../../mocks/dcasclient.gen.go --fake-name DCASClient github.com/trustbloc/fabric-peer-ext/pkg/collections/offledger/dcas/client.DCAS
//go:generate counterfeiter -o ./../../mocks/dcasprovider.gen.go --fake-name DCASClientProvider . DCASClientProvider
//go:generate counterfeiter -o ./../mocks/offledgerprovider.gen.go --fake-name OffLedgerClientProvider . OffLedgerClientProvider
//go:generate counterfeiter -o ./../mocks/bcclientprovider.gen.go --fake-name BlockchainClientProvider . BlockchainClientProvider
//go:generate counterfeiter -o ./../mocks/bcclient.gen.go --fake-name BlockchainClient ../../client Blockchain

// DCASClientProvider is a DCAS client provider
type DCASClientProvider interface {
	ForChannel(channelID string) (dcasclient.DCAS, error)
}

// OffLedgerClientProvider is an off-ledger client provider
type OffLedgerClientProvider interface {
	ForChannel(channelID string) (client.OffLedger, error)
}

// BlockchainClientProvider provides a blockchain client for a given channel
type BlockchainClientProvider interface {
	ForChannel(channelID string) (bcclient.Blockchain, error)
}
