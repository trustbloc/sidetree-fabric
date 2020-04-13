/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package dcashandler

type retrieveError struct {
	status int
	code   ResultCode
}

func newRetrieveError(status int, code ResultCode) *retrieveError {
	return &retrieveError{
		status: status,
		code:   code,
	}
}

// Error returns the error string
func (e *retrieveError) Error() string {
	return e.code
}

// Status returns the status code
func (e *retrieveError) Status() int {
	return e.status
}

// ResultCode returns the result code
func (e *retrieveError) ResultCode() ResultCode {
	return e.code
}
