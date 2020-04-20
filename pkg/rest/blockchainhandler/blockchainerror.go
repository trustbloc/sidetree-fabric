/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package blockchainhandler

import (
	"net/http"
)

type blockchainError struct {
	status int
	code   ResultCode
}

func newBadRequestError(code ResultCode) *blockchainError {
	return &blockchainError{
		status: http.StatusBadRequest,
		code:   code,
	}
}

// Error returns the error string
func (e *blockchainError) Error() string {
	return e.code
}

// Status returns the status code
func (e *blockchainError) Status() int {
	return e.status
}

// ResultCode returns the result code
func (e *blockchainError) ResultCode() ResultCode {
	return e.code
}
