/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package blockchain

import (
	"sync"

	"github.com/pkg/errors"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
)

const (
	sidetreeTxnCC  = "sidetreetxn_cc"
	writeAnchorFcn = "writeAnchor"
)

// Client implements blockchain client for writing anchors
type Client struct {
	lock            sync.RWMutex
	channelProvider context.ChannelProvider
	channelClient   chClient
}

type chClient interface {
	Execute(request channel.Request, options ...channel.RequestOption) (channel.Response, error)
}

// New returns a new blockchain client
func New(channelProvider context.ChannelProvider) *Client {
	return &Client{channelProvider: channelProvider}
}

// WriteAnchor writes anchor file address to blockchain
func (c *Client) WriteAnchor(anchor string) error {

	client, err := c.getClient()
	if err != nil {
		return errors.Wrap(err, "failed to get channel client")
	}

	_, err = client.Execute(channel.Request{
		ChaincodeID: sidetreeTxnCC,
		Fcn:         writeAnchorFcn,
		Args:        [][]byte{[]byte(anchor)},
	})

	if err != nil {
		return errors.Wrap(err, "failed to store anchor file address")
	}

	return nil
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
