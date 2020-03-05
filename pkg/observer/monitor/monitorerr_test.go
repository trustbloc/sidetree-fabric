/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package monitor

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMonitorError(t *testing.T) {
	err := errors.New("cause of error")
	merr := newMonitorError(err, true)
	require.NotNil(t, merr)
	require.Equal(t, err.Error(), merr.Error())
	require.True(t, merr.Transient())
}
