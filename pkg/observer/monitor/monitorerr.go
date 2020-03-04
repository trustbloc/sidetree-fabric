/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package monitor

type monitorError struct {
	cause     error
	transient bool
}

func newMonitorError(cause error, transient bool) monitorError {
	return monitorError{
		cause:     cause,
		transient: transient,
	}
}

func (e monitorError) Error() string {
	return e.cause.Error()
}

func (e monitorError) Transient() bool {
	return e.transient
}
