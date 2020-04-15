/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package blockchainhandler

// Config defines the configuration for a DCAS handler
type Config struct {
	// Version is the version of the blockchain endpoint
	Version string
	// BasePath is the base context path of the REST endpoint
	BasePath string
}
