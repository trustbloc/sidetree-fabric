/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package blockchainhandler

// TimeResponse contains the response from the /time request
type TimeResponse struct {
	Time string `json:"time"`
	Hash []byte `json:"hash"`
}
