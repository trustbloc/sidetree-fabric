/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package httpserver

import (
	"net/http"
)

// StatusMsg is a descriptive message that's returned in the response
type StatusMsg = string

const (
	// StatusBadRequest indicate that the request was invalid
	StatusBadRequest StatusMsg = "bad-request"
	// StatusEmptyContent indicates that no content was provided in the POST
	StatusEmptyContent StatusMsg = "empty-content"
	// StatusNotFound indicates that the content for the provided hash was not found
	StatusNotFound StatusMsg = "not-found"
	// StatusServerError indicates that the server experienced an unexpected error
	StatusServerError StatusMsg = "server-error"
)

var (
	// ServerError is an internal server error
	ServerError = NewError(http.StatusInternalServerError, StatusServerError)
	// NotFoundError indicates that the requested content was not found
	NotFoundError = NewError(http.StatusNotFound, StatusNotFound)
	// BadRequestError indicates that the request is invalid
	BadRequestError = NewError(http.StatusBadRequest, StatusBadRequest)
)

// Error holds additional context associated with the HTTP request
type Error struct {
	status int
	msg    StatusMsg
}

// NewError returns a new Error
func NewError(status int, msg StatusMsg) *Error {
	return &Error{
		status: status,
		msg:    msg,
	}
}

// Error returns the error string
func (e *Error) Error() string {
	return e.msg
}

// Status returns the status code
func (e *Error) Status() int {
	return e.status
}

// StatusMsg returns the status message
func (e *Error) StatusMsg() StatusMsg {
	return e.msg
}
