/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package sidetreehandler

import (
	"github.com/trustbloc/sidetree-fabric/pkg/rest/authhandler"
)

// DocumentType specifies the type of the Sidetree document
type DocumentType = string

const (
	// DIDDocType indicates that the document is a DID document (this is the default)
	DIDDocType DocumentType = ""
	// FileIndexType indicates that the document contains a file index
	FileIndexType DocumentType = "FILE_INDEX"
)

// Config holds Sidetree endpoint handler config
type Config struct {
	Authorization authhandler.Config

	// DocType specifies the document type (DID or File index)
	DocType DocumentType
	// Namespace is the namespace prefix used in the ID of the document
	Namespace string
	// BasePath is the base context path of the REST endpoint
	BasePath string
}
