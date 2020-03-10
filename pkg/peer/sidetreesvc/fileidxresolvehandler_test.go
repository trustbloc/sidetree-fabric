/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package sidetreesvc

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFileIdxResolveHandler(t *testing.T) {
	path := "/file"

	h := newFileIdxResolveHandler(path, nil)
	require.NotNil(t, h)
	require.Equal(t, "/file/{id}", h.Path())
	require.NotNil(t, h.Handler())
	require.Equal(t, http.MethodGet, h.Method())
}
