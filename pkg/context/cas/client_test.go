/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package cas

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"

	"github.com/trustbloc/sidetree-fabric/pkg/mocks"
)

func TestNew(t *testing.T) {
	c := New(&mocks.DCASClient{})
	require.NotNil(t, c)
}

func TestWriteContent(t *testing.T) {
	content := []byte("content")

	dcasClient := mocks.NewDCASClient()

	cas := New(dcasClient)
	require.NotNil(t, cas)

	address, err := cas.Write(content)
	require.Nil(t, err)
	require.NotEmpty(t, address)

	read, err := cas.Read(address)
	require.Nil(t, err)
	require.NotNil(t, read)
	require.Equal(t, content, read)
}

func TestWriteContentError(t *testing.T) {
	testErr := errors.New("channel error")

	dcasClient := &mocks.DCASClient{}
	dcasClient.PutReturns("", testErr)

	cas := New(dcasClient)

	content := []byte("content")
	address, err := cas.Write(content)
	require.NotNil(t, err)
	require.Empty(t, address)
	require.Contains(t, err.Error(), testErr.Error())
}

func TestReadContentError(t *testing.T) {
	dcasClient := mocks.NewDCASClient()

	cas := New(dcasClient)

	t.Run("Error", func(t *testing.T) {
		testErr := errors.New("channel error")
		dcasClient.WithGetError(testErr)
		defer dcasClient.WithGetError(nil)

		read, err := cas.Read("address")
		require.Error(t, err)
		require.Nil(t, read)
		require.Contains(t, err.Error(), testErr.Error())
	})

	t.Run("Error", func(t *testing.T) {
		read, err := cas.Read("address")
		require.Error(t, err)
		require.Nil(t, read)
		require.Contains(t, err.Error(), "not found")
	})
}
