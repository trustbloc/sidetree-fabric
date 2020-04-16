/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package versionhandler

// Response contains the response from a version request
type Response struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}
