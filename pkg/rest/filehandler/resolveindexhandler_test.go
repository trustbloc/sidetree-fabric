/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package filehandler

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolveIndex(t *testing.T) {
	path := "/file"

	h := NewResolveIndexHandler(path, nil)
	require.NotNil(t, h)
	require.Equal(t, "/file/identifiers/{id}", h.Path())
	require.NotNil(t, h.Handler())
	require.Equal(t, http.MethodGet, h.Method())
}
