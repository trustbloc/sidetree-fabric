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

	"github.com/sirupsen/logrus"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
)

const (
	documentCC = "document_cc"
	queryFcn   = "queryByID"
)

var logger = logrus.New()

// Client implements client for accessing document operations
type Client struct {
	lock            sync.RWMutex
	channelProvider context.ChannelProvider
	channelClient   chClient
	namespace       string
}

type chClient interface {
	Query(request channel.Request, options ...channel.RequestOption) (channel.Response, error)
}

// New returns a new store client
func New(channelProvider context.ChannelProvider, namespace string) *Client {
	return &Client{channelProvider: channelProvider, namespace: namespace}
}

// Get retrieves all document operations for specified document ID
func (c *Client) Get(uniqueSuffix string) ([]batch.Operation, error) {

	client, err := c.getClient()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get channel client")
	}

	id := c.namespace + uniqueSuffix
	logger.Debugf("get operations for doc[%s]", id)

	response, err := client.Query(channel.Request{
		ChaincodeID: documentCC,
		Fcn:         queryFcn,
		Args:        [][]byte{[]byte(id)},
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to query document operations")
	}

	return getOperations(response.Payload)
}

func getOperations(payload []byte) ([]batch.Operation, error) {

	var ops [][]byte
	err := json.Unmarshal(payload, &ops)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal operations: %s", err)
	}

	var operations []batch.Operation
	for _, opBytes := range ops {
		var op batch.Operation
		err = json.Unmarshal(opBytes, &op)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to unmarshal operation")
		}
		operations = append(operations, op)
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
