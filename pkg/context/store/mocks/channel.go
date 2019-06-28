/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"encoding/json"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
)

// MockChannelClient mocks channel client
type MockChannelClient struct {
	Err error
}

// NewMockChannelClient returns mock channel client
func NewMockChannelClient() *MockChannelClient {
	return &MockChannelClient{}
}

// Query mocks query
func (cc *MockChannelClient) Query(request channel.Request, options ...channel.RequestOption) (channel.Response, error) {
	if cc.Err != nil {
		return channel.Response{}, cc.Err
	}

	return channel.Response{Payload: getDefaultOperations()}, nil
}

func getJSON(ops [][]byte) []byte {

	bytes, err := json.Marshal(ops)
	if err != nil {
		panic(err)
	}

	return bytes
}

// Operation defines sample operation
type Operation struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

func getDefaultOperations() []byte {

	var ops [][]byte

	bytes, err := json.Marshal(Operation{ID: "did:sidetree:abc", Type: "create"})
	if err != nil {
		panic(err)
	}

	ops = append(ops, bytes)

	return getJSON(ops)
}
