/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package dcashandler

// ResultCode specifies the status string of a read result
type ResultCode = string

const (
	// CodeCasNotReachable indicates that there was an error when attempting to communicate with the CAS service
	CodeCasNotReachable ResultCode = "cas_not_reachable"
	// CodeInvalidHash indicates that the hash was not specified or is invalid
	CodeInvalidHash ResultCode = "content_hash_invalid"
	// CodeMaxSizeExceeded indicates that the resulting content exceeds the maximum size specified in the request
	CodeMaxSizeExceeded ResultCode = "content_exceeds_maximum_allowed_size"
	// CodeMaxSizeNotSpecified indicates that the max-size parameter was not specified in the request
	CodeMaxSizeNotSpecified ResultCode = "content_max_size_not_specified"
	// CodeNotFound indicates that the content for the given hash was not found
	CodeNotFound ResultCode = "content_not_found"
)

// UploadResponse contains the response from a CAS upload
type UploadResponse struct {
	Hash string `json:"hash"`
}
