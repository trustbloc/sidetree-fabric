/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package config

import (
	"time"

	protocolApi "github.com/trustbloc/sidetree-core-go/pkg/api/protocol"

	"github.com/trustbloc/sidetree-fabric/pkg/filehandler"
	"github.com/trustbloc/sidetree-fabric/pkg/rest/blockchainhandler"
	"github.com/trustbloc/sidetree-fabric/pkg/rest/dcashandler"
)

// DocumentType specifies the type of the Sidetree document
type DocumentType = string

const (
	// DIDDocType indicates that the document is a DID document (this is the default)
	DIDDocType DocumentType = ""
	// FileIndexType indicates that the document contains a file index
	FileIndexType DocumentType = "FILE_INDEX"
)

// Namespace holds Sidetree namespace config
type Namespace struct {
	DocType   DocumentType
	Namespace string
	BasePath  string
}

// Monitor holds Sidetree monitor config
type Monitor struct {
	Period                time.Duration
	MetaDataChaincodeName string
}

// SidetreePeer holds peer-specific Sidetree config
type SidetreePeer struct {
	Monitor    Monitor
	Namespaces []Namespace
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
	LoadFileHandlers(mspID, peerID string) ([]filehandler.Config, error)
	LoadDCASHandlers(mspID, peerID string) ([]dcashandler.Config, error)
	LoadBlockchainHandlers(mspID, peerID string) ([]blockchainhandler.Config, error)
	LoadDCAS() (DCAS, error)
}
