/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package filehandler

// Config defines the configuration for a file handler
type Config struct {
	// BasePath is the base context path of the REST endpoint
	BasePath string
	// ChaincodeName is the name of the chaincode that stores the files
	ChaincodeName string
	// Collection is the name of the DCAS collection that stores the file
	Collection string
	// IndexNamespace is the namespace of the index Sidetree document
	IndexNamespace string
	// IndexDocID is ID of the Sidetree document that contains the index of all files stored in this path
	IndexDocID string
}
