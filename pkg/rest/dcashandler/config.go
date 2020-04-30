/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package dcashandler

import (
	"github.com/trustbloc/sidetree-fabric/pkg/rest/authhandler"
)

// Config defines the configuration for a DCAS handler
type Config struct {
	Authorization authhandler.Config

	// Version is the version of the DCAS endpoint
	Version string
	// BasePath is the base context path of the REST endpoint
	BasePath string
	// ChaincodeName is the name of the chaincode that stores the content
	ChaincodeName string
	// Collection is the name of the DCAS collection that stores the content
	Collection string
}
