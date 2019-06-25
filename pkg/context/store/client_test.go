/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package store

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/trustbloc/sidetree-fabric/pkg/context/store/mocks"

	"github.com/pkg/errors"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	fabMocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
)

const (
	chID = "mychannel"
	id   = "id"
)

func TestNew(t *testing.T) {
	ctx := channelProvider(chID)
	c := New(ctx)
	require.NotNil(t, c)
}

func TestGetClientError(t *testing.T) {
	testErr := errors.New("provider error")
	ctx := channelProviderWithError(testErr)

	c := New(ctx)
	require.NotNil(t, c)

	payload, err := c.Get(id)
	require.NotNil(t, err)
	require.Nil(t, payload)
	require.Contains(t, err.Error(), testErr.Error())
}

func TestWriteContent(t *testing.T) {
	c := New(channelProvider(chID))

	c.channelClient = mocks.NewMockChannelClient()

	read, err := c.Get(id)
	require.Nil(t, err)
	require.NotNil(t, read)
}

func TestReadContentError(t *testing.T) {

	testErr := errors.New("channel error")
	cc := mocks.NewMockChannelClient()
	cc.Err = testErr

	c := New(channelProvider(chID))
	c.channelClient = cc

	read, err := c.Get(id)
	require.NotNil(t, err)
	require.Nil(t, read)
	require.Contains(t, err.Error(), testErr.Error())
}

func channelProvider(channelID string) context.ChannelProvider {
	channelProvider := func() (context.Channel, error) {
		return fabMocks.NewMockChannel(channelID)
	}
	return channelProvider
}

func channelProviderWithError(err error) context.ChannelProvider {
	channelProvider := func() (context.Channel, error) {
		return nil, err
	}
	return channelProvider
}
