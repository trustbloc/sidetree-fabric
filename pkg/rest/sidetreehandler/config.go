/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package sidetreehandler

import (
	"time"

	"github.com/trustbloc/sidetree-fabric/pkg/common"
	"github.com/trustbloc/sidetree-fabric/pkg/rest/authhandler"
)

// Config holds Sidetree endpoint handler config
type Config struct {
	Authorization authhandler.Config

	// Version contains the version of the Sidetree endpoint
	Version string
	// DocType specifies the document type (DID or File index)
	DocType common.DocumentType
	// Namespace is the namespace prefix used in the ID of the document
	Namespace string
	// Aliases is the namespace aliases that can be used in the ID of the document
	Aliases []string
	// BasePath is the base context path of the REST endpoint
	BasePath string
	// DocumentCacheSize is the maximum number of documents to hold in the cache. If 0 then the default cache size is used.
	DocumentCacheSize uint
	// DocumentExpiry returns the expiration time of a cached document. If zero then the document never expires.
	DocumentExpiry time.Duration
}
