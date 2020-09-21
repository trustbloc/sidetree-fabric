/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package protocolversion

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRegisterFactories(t *testing.T) {
	require.NotPanics(t, RegisterFactories)
}
