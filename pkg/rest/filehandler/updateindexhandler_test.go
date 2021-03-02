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

func TestUpdateIndex(t *testing.T) {
	path := "/file/operations"

	h := NewUpdateIndexHandler(path, nil, nil)
	require.NotNil(t, h)
	require.Equal(t, path, h.Path())
	require.NotNil(t, h.Handler())
	require.Equal(t, http.MethodPost, h.Method())
}
