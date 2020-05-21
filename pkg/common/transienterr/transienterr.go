/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package transienterr

import (
	"fmt"

	"github.com/pkg/errors"
)

// Code is the error code
type Code string

const (
	// CodeUnknown is an unspecified error code
	CodeUnknown Code = "UNKNOWN"
	// CodeNotFound indicates that an item was not found
	CodeNotFound Code = "NOT_FOUND"
	// CodeBlockchain indicates that an error occurred while committing a transaction to the blockchain
	CodeBlockchain Code = "BLOCKCHAIN"
	// CodeDB indicates that an error occurred while accessing the database
	CodeDB Code = "DB"
)

// Error is a transient error, meaning that a retry on the request may succeed
type Error struct {
	error
	code Code
}

// New returns a transient error which wraps the given error
// and assigns the given code
func New(cause error, code Code) Error {
	return Error{
		error: cause,
		code:  code,
	}
}

// String returns the string representation of the error
func (err *Error) String() string {
	return fmt.Sprintf("%s - Code: %s", err.Error(), err.code)
}

// Is returns true if the given error is a transient error
func Is(err error) bool {
	_, ok := errors.Cause(err).(Error)
	return ok
}

// GetCode returns true if the given error is a transient error
func GetCode(err error) Code {
	terr, ok := errors.Cause(err).(Error)
	if !ok {
		return CodeUnknown
	}

	return terr.code
}

// Code is the error code
func (err *Error) Code() Code {
	return err.code
}
