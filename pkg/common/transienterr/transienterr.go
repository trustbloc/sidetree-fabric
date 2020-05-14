/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package transienterr

import (
	"github.com/pkg/errors"
)

// Error is a transient error, meaning that a retry on the request may succeed
type Error struct {
	error
}

// New returns a transient error which wraps the given error
func New(cause error) Error {
	return Error{error: cause}
}

// Is returns true if the given error is a transient error
func Is(err error) bool {
	_, ok := errors.Cause(err).(Error)
	return ok
}
