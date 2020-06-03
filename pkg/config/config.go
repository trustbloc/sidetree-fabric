/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package config

import (
	"time"

	protocolApi "github.com/trustbloc/sidetree-core-go/pkg/api/protocol"

	"github.com/trustbloc/sidetree-fabric/pkg/rest/blockchainhandler"
	"github.com/trustbloc/sidetree-fabric/pkg/rest/dcashandler"
	"github.com/trustbloc/sidetree-fabric/pkg/rest/discoveryhandler"
	"github.com/trustbloc/sidetree-fabric/pkg/rest/filehandler"
	"github.com/trustbloc/sidetree-fabric/pkg/rest/sidetreehandler"
)

// Observer holds Sidetree Observer config
type Observer struct {
	// Period is the scheduled period for processing blocks
	Period time.Duration
	// MetaDataChaincodeName is the name of the chaincode that stores metadata
	MetaDataChaincodeName string
	// MaxAttempts is the maximum number of attempts to process a transaction. When a transient error
	// occurs then a retry is attempted at the next scheduled interval. After processing has failed
	// MaxAttempts times, the batch is lost and processing continues at the next transaction in the block.
	MaxAttempts int
}

// SidetreePeer holds peer-specific Sidetree config
type SidetreePeer struct {
	Observer Observer
}

// DCAS holds Distributed Content Addressable Store (DCAS) configuration
type DCAS struct {
	ChaincodeName string
	Collection    string
}

// Sidetree holds general Sidetree configuration
type Sidetree struct {
	ChaincodeName      string
	Collection         string
	BatchWriterTimeout time.Duration
}

// SidetreeService is a service that loads Sidetree configuration
type SidetreeService interface {
	LoadProtocols(namespace string) (map[string]protocolApi.Protocol, error)
	LoadSidetree(namespace string) (Sidetree, error)
	LoadSidetreePeer(mspID, peerID string) (SidetreePeer, error)
	LoadSidetreeHandlers(mspID, peerID string) ([]sidetreehandler.Config, error)
	LoadFileHandlers(mspID, peerID string) ([]filehandler.Config, error)
	LoadDCASHandlers(mspID, peerID string) ([]dcashandler.Config, error)
	LoadBlockchainHandlers(mspID, peerID string) ([]blockchainhandler.Config, error)
	LoadDiscoveryHandlers(mspID, peerID string) ([]discoveryhandler.Config, error)
	LoadDCAS() (DCAS, error)
}
