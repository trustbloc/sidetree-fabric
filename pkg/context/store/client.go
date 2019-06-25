/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package store

import (
	"fmt"
	"sync"

	"github.com/trustbloc/sidetree-core-go/pkg/api/batch"

	"github.com/pkg/errors"

	"encoding/json"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
)

const (
	documentCC = "document_cc"
	queryFcn   = "queryByID"
)

// Client implements client for accessing document operations
type Client struct {
	lock            sync.RWMutex
	channelProvider context.ChannelProvider
	channelClient   chClient
}

type chClient interface {
	Query(request channel.Request, options ...channel.RequestOption) (channel.Response, error)
}

// New returns a new store client
func New(channelProvider context.ChannelProvider) *Client {
	return &Client{channelProvider: channelProvider}
}

// Get retrieves all document operations for specified document ID
func (c *Client) Get(id string) ([]batch.Operation, error) {

	client, err := c.getClient()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get channel client")
	}

	response, err := client.Query(channel.Request{
		ChaincodeID: documentCC,
		Fcn:         queryFcn,
		Args:        [][]byte{[]byte(id)},
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to query document operations")
	}

	var operations []batch.Operation
	err = json.Unmarshal(response.Payload, &operations)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal operations: %s", err)
	}

	return operations, nil
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
