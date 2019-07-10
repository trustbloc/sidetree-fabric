/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package blockchain

import (
	"testing"

	"github.com/trustbloc/sidetree-fabric/pkg/context/blockchain/mocks"

	"github.com/pkg/errors"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/stretchr/testify/require"

	fabMocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
)

const chID = "mychannel"

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

	err := c.WriteAnchor("anchor")
	require.NotNil(t, err)
	require.Contains(t, err.Error(), testErr.Error())
}

func TestWriteAnchor(t *testing.T) {
	c := New(channelProvider(chID))

	c.channelClient = mocks.NewMockChannelClient()

	err := c.WriteAnchor("anchor")
	require.Nil(t, err)
}

func TestWriteAnchorError(t *testing.T) {

	testErr := errors.New("channel error")
	cc := mocks.NewMockChannelClient()
	cc.Err = testErr

	bc := New(channelProvider(chID))
	bc.channelClient = cc

	err := bc.WriteAnchor("anchor")
	require.NotNil(t, err)
	require.Contains(t, err.Error(), testErr.Error())
}

func TestClient_Read(t *testing.T) {
	require.PanicsWithValue(t, "not implemented", func() {
		c := New(channelProvider(chID))
		c.channelClient = mocks.NewMockChannelClient()
		c.Read(1000)
	})
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
