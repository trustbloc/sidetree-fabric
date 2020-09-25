/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package common

// DocumentType specifies the type of the Sidetree document
type DocumentType = string

const (
	// DIDDocType indicates that the document is a DID document (this is the default)
	DIDDocType DocumentType = ""
	// FileIndexType indicates that the document contains a file index
	FileIndexType DocumentType = "FILE_INDEX"
)
