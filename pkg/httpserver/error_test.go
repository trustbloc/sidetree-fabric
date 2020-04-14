/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package httpserver

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestError(t *testing.T) {
	err := NewError(http.StatusBadRequest, StatusBadRequest)
	require.EqualError(t, err, StatusBadRequest)
	require.Equal(t, http.StatusBadRequest, err.Status())
	require.Equal(t, StatusBadRequest, err.StatusMsg())

}
