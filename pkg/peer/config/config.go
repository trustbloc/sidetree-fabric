/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package config

import (
	"time"
)

const (
	// GlobalMSPID is used as the consortium-wide MSP ID (i.e. non org-specific)
	GlobalMSPID = "general"

	// SidetreeAppVersion is the version of the Sidetree config application
	SidetreeAppVersion = "1"

	// ProtocolComponentName is the name of the Sidetree protocol config component
	ProtocolComponentName = "protocol"

	// SidetreePeerAppName is the name of the Sidetree config application
	SidetreePeerAppName = "sidetree"

	// SidetreePeerAppVersion is the version of the Sidetree config application
	SidetreePeerAppVersion = "1"

	// FileHandlerAppName is the name of the file handler config application
	FileHandlerAppName = "file-handler"

	// FileHandlerAppVersion is the version of the file handler config application
	FileHandlerAppVersion = "1"
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
	Period time.Duration
}

// SidetreePeer holds peer-specific Sidetree config
type SidetreePeer struct {
	Monitor    Monitor
	Namespaces []Namespace
}

// Sidetree holds general Sidetree configuration
type Sidetree struct {
	BatchWriterTimeout time.Duration
}
