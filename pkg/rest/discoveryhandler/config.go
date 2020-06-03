/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package discoveryhandler

import "github.com/trustbloc/sidetree-fabric/pkg/rest/authhandler"

// Config defines the configuration for a blockchain handler
type Config struct {
	Authorization authhandler.Config

	// BasePath is the base context path of the REST endpoint
	BasePath string
}
