/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package cas

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	stmocks "github.com/trustbloc/sidetree-fabric/pkg/mocks"
)

const chID = "mychannel"

func TestNew(t *testing.T) {
	dcasProvider := &stmocks.DCASClientProvider{}
	c := New(chID, dcasProvider)
	require.NotNil(t, c)
}

func TestForChannelError(t *testing.T) {
	testErr := errors.New("provider error")

	dcasProvider := &stmocks.DCASClientProvider{}
	dcasProvider.ForChannelReturns(nil, testErr)

	c := New(chID, dcasProvider)
	require.NotNil(t, c)

	content := []byte("content")
	address, err := c.Write(content)
	require.NotNil(t, err)
	require.Empty(t, address)
	require.Contains(t, err.Error(), testErr.Error())

	payload, err := c.Read("address")
	require.EqualError(t, err, testErr.Error())
	require.Empty(t, payload)
}

func TestWriteContent(t *testing.T) {
	content := []byte("content")

	dcasClient := &stmocks.DCASClient{}
	dcasClient.PutReturns("address", nil)
	dcasClient.GetReturns(content, nil)

	dcasProvider := &stmocks.DCASClientProvider{}
	dcasProvider.ForChannelReturns(dcasClient, nil)

	cas := New(chID, dcasProvider)
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

	dcasClient := &stmocks.DCASClient{}
	dcasClient.PutReturns("", testErr)
	dcasProvider := &stmocks.DCASClientProvider{}
	dcasProvider.ForChannelReturns(dcasClient, nil)

	cas := New(chID, dcasProvider)

	content := []byte("content")
	address, err := cas.Write(content)
	require.NotNil(t, err)
	require.Empty(t, address)
	require.Contains(t, err.Error(), testErr.Error())
}

func TestReadContentError(t *testing.T) {

	testErr := errors.New("channel error")

	dcasClient := &stmocks.DCASClient{}
	dcasClient.GetReturns(nil, testErr)
	dcasProvider := &stmocks.DCASClientProvider{}
	dcasProvider.ForChannelReturns(dcasClient, nil)

	cas := New(chID, dcasProvider)

	read, err := cas.Read("address")
	require.NotNil(t, err)
	require.Nil(t, read)
	require.Contains(t, err.Error(), testErr.Error())
}
