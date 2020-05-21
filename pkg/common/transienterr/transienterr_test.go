/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package transienterr

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestError(t *testing.T) {
	cause := errors.New("cause of error")
	err := New(cause, CodeNotFound)
	require.NotNil(t, err)
	require.Equal(t, cause.Error(), err.Error())
	require.Equal(t, CodeNotFound, err.Code())
	require.Equal(t, CodeNotFound, GetCode(err))
	require.Equal(t, "cause of error - Code: NOT_FOUND", err.String())
}

func TestIs(t *testing.T) {
	cause := errors.New("cause of error")
	err := New(cause, CodeDB)
	require.NotNil(t, err)
	require.True(t, Is(err))
	require.Equal(t, CodeDB, GetCode(err))
	require.False(t, Is(cause))
	require.Equal(t, CodeUnknown, GetCode(cause))
}
