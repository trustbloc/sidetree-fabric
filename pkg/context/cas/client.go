/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package cas

import (
	"sync"

	"github.com/pkg/errors"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
)

const (
	sidetreeTxnCC = "sidetreetxn_cc"
	writeFcn      = "writeContent"
	readFcn       = "readContent"
)

// Client implements client for accessing the underlying content addressable storage
type Client struct {
	lock            sync.RWMutex
	channelProvider context.ChannelProvider
	channelClient   chClient
}

type chClient interface {
	Query(request channel.Request, options ...channel.RequestOption) (channel.Response, error)
	Execute(request channel.Request, options ...channel.RequestOption) (channel.Response, error)
}

// New returns a new CAS client
func New(channelProvider context.ChannelProvider) *Client {

	return &Client{channelProvider: channelProvider}
}

// Write writes the given content to content addressable storage
// returns the SHA256 hash in base64url encoding which represents the address of the content.
func (c *Client) Write(content []byte) (string, error) {

	client, err := c.getClient()
	if err != nil {
		return "", errors.Wrap(err, "failed to get channel client")
	}

	response, err := client.Execute(channel.Request{
		ChaincodeID: sidetreeTxnCC,
		Fcn:         writeFcn,
		Args:        [][]byte{content},
	})

	if err != nil {
		return "", errors.Wrap(err, "failed to store content")
	}

	return string(response.Payload), nil
}

// Read reads the content at the given address from content addressable storage
// returns the content of the given address.
func (c *Client) Read(address string) ([]byte, error) {

	client, err := c.getClient()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get channel client")
	}

	response, err := client.Query(channel.Request{
		ChaincodeID: sidetreeTxnCC,
		Fcn:         readFcn,
		Args:        [][]byte{[]byte(address)},
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to read content at requested address")
	}

	return response.Payload, nil
}

func (c *Client) getClient() (chClient, error) {

	c.lock.RLock()
	chc := c.channelClient
	c.lock.RUnlock()

	if chc != nil {
		return chc, nil
	}

	c.lock.Lock()
	defer c.lock.Unlock()

	if c.channelClient == nil {
		channelClient, err := channel.New(c.channelProvider)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create channel client")
		}

		c.channelClient = channelClient
	}

	return c.channelClient, nil
}
