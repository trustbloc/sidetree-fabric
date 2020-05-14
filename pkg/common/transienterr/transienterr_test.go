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
	err := New(cause)
	require.NotNil(t, err)
	require.Equal(t, cause.Error(), err.Error())
}

func TestIs(t *testing.T) {
	cause := errors.New("cause of error")
	err := New(cause)
	require.NotNil(t, err)
	require.True(t, Is(err))
	require.False(t, Is(cause))
}
