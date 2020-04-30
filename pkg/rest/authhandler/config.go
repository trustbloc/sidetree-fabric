/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package authhandler

// Config contains the config for endpoint authorization
type Config struct {
	// ReadTokens contains a set of names of tokens for authorizing read requests
	ReadTokens []string
	// WriteTokens contains a set of names of tokens for authorizing write requests
	WriteTokens []string
}
