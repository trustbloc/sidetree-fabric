/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package sidetreehandler

import (
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
	// BasePath is the base context path of the REST endpoint
	BasePath string
}
