/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package cas

import (
	"testing"

	"github.com/trustbloc/sidetree-fabric/pkg/context/cas/mocks"

	"github.com/pkg/errors"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	fabMocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	"github.com/stretchr/testify/assert"
)

const chID = "mychannel"

func TestNew(t *testing.T) {
	ctx := channelProvider(chID)
	c := New(ctx)
	assert.NotNil(t, c)
}

func TestGetClientError(t *testing.T) {
	testErr := errors.New("provider error")
	ctx := channelProviderWithError(testErr)

	c := New(ctx)
	assert.NotNil(t, c)

	content := []byte("content")
	address, err := c.Write(content)
	assert.NotNil(t, err)
	assert.Empty(t, address)
	assert.Contains(t, err.Error(), testErr.Error())


	payload, err := c.Read("address")
	assert.NotNil(t, err)
	assert.Nil(t, payload)
	assert.Contains(t, err.Error(), testErr.Error())
}

func TestWriteContent(t *testing.T) {
	cas := New(channelProvider(chID))

	cas.channelClient = mocks.NewMockChannelClient()

	content := []byte("content")
	address, err := cas.Write(content)
	assert.Nil(t, err)
	assert.NotEmpty(t, address)

	read, err := cas.Read(address)
	assert.Nil(t, err)
	assert.NotNil(t, read)
	assert.Equal(t, content, read)
}

func TestWriteContentError(t *testing.T) {

	testErr := errors.New("channel error")
	cc := mocks.NewMockChannelClient()
	cc.Err = testErr

	cas := New(channelProvider(chID))
	cas.channelClient = cc

	content := []byte("content")
	address, err := cas.Write(content)
	assert.NotNil(t, err)
	assert.Empty(t, address)
	assert.Contains(t, err.Error(), testErr.Error())
}

func TestReadContentError(t *testing.T) {

	testErr := errors.New("channel error")
	cc := mocks.NewMockChannelClient()
	cc.Err = testErr

	cas := New(channelProvider(chID))
	cas.channelClient = cc

	read, err := cas.Read("address")
	assert.NotNil(t, err)
	assert.Nil(t, read)
	assert.Contains(t, err.Error(), testErr.Error())
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
