/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
package cas

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/trustbloc/fabric-peer-ext/pkg/collections/offledger/dcas"
	"github.com/trustbloc/sidetree-core-go/pkg/docutil"

	"github.com/trustbloc/sidetree-fabric/pkg/mocks"
)

func TestWrite(t *testing.T) {
	casClient := mocks.NewDCASClient()
	client := New(casClient)

	t.Run("Success", func(t *testing.T) {
		content := getOperationBytes(getCreateOperation())
		addr, err := client.Write(content)
		require.Nil(t, err)
		require.NotEmpty(t, addr)
		require.NoError(t, dcas.ValidateCID(addr))

		payload, err := client.Read(addr)
		require.Nil(t, err)
		require.NotNil(t, payload)
		require.Equal(t, content, payload)

		// test write same content
		addr2, err := client.Write(content)
		require.Nil(t, err)
		require.Equal(t, addr, addr2)
	})

	t.Run("Write error", func(t *testing.T) {
		casClient.WithPutError(errors.New("injected error"))
		defer casClient.WithPutError(nil)

		content := getOperationBytes(getCreateOperation())
		addr, err := client.Write(content)
		require.Error(t, err)
		require.Empty(t, addr)
	})
}

func TestRead(t *testing.T) {
	casClient := mocks.NewDCASClient()
	client := New(casClient)

	content := getOperationBytes(getCreateOperation())
	addr, err := client.Write(content)
	require.Nil(t, err)
	require.NotNil(t, addr)

	read, err := client.Read(addr)
	require.Nil(t, err)
	require.NotNil(t, read)
	require.Equal(t, read, content)

	read, err = client.Read("non-existent")
	require.Nil(t, err)
	require.Nil(t, read)

	testErr := errors.New("read error")
	casClient.WithGetError(testErr)

	payload, err := client.Read("address")
	require.Error(t, err)
	require.Nil(t, payload)
	require.Contains(t, err.Error(), testErr.Error())
}

func getOperationBytes(op *Operation) []byte {

	bytes, err := docutil.MarshalCanonical(op)
	if err != nil {
		panic(err)
	}

	return bytes
}

func getCreateOperation() *Operation {
	return &Operation{UniqueSuffix: "abc", Type: "create"}
}

// Operation defines sample operation
type Operation struct {
	Type         string `json:"type"`
	UniqueSuffix string `json:"uniqueSuffix"`
}
