/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package dcashandler

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRetrieveError(t *testing.T) {
	err := newRetrieveError(http.StatusBadRequest, CodeInvalidHash)
	require.EqualError(t, err, CodeInvalidHash)
	require.Equal(t, http.StatusBadRequest, err.Status())
	require.Equal(t, CodeInvalidHash, err.ResultCode())
}
