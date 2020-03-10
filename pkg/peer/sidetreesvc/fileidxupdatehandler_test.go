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

func TestFileIdxUpdateHandler(t *testing.T) {
	path := "/file"

	h := newFileIdxUpdateHandler(path, nil)
	require.NotNil(t, h)
	require.Equal(t, "/file", h.Path())
	require.NotNil(t, h.Handler())
	require.Equal(t, http.MethodPost, h.Method())
}
