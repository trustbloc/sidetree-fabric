/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package common

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMarshalDCAS(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		o := &testObject{
			Field1: "field1",
			Field2: 1000,
		}

		key, bytes, err := MarshalDCAS(o)
		require.NoError(t, err)
		require.NotEmpty(t, key)
		require.NotEmpty(t, bytes)
	})

	t.Run("Error", func(t *testing.T) {
		key, bytes, err := MarshalDCAS(testFunc)
		require.Errorf(t, err, "expecting error marshalling function")
		require.Empty(t, key)
		require.Empty(t, bytes)
	})
}

type testObject struct {
	Field1 string
	Field2 int
}

func testFunc() {}
