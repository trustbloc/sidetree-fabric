/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"github.com/trustbloc/sidetree-core-go/pkg/mocks"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
)

// MockChannelClient mocks channel client
type MockChannelClient struct {
	Err error
	cas *mocks.MockCasClient
}

// NewMockChannelClient returns mock channel client
func NewMockChannelClient() *MockChannelClient {
	return &MockChannelClient{cas: mocks.NewMockCasClient(nil)}
}

// Query mocks query
func (cc *MockChannelClient) Query(request channel.Request, options ...channel.RequestOption) (channel.Response, error) {
	if cc.Err != nil {
		return channel.Response{}, cc.Err
	}

	content, err := cc.cas.Read(string(request.Args[0]))
	if err != nil {
		return channel.Response{}, err
	}

	return channel.Response{Payload: content}, nil
}

// Execute mocks execute
func (cc *MockChannelClient) Execute(request channel.Request, options ...channel.RequestOption) (channel.Response, error) {

	if cc.Err != nil {
		return channel.Response{}, cc.Err
	}

	address, err := cc.cas.Write(request.Args[0])
	if err != nil {
		return channel.Response{}, err
	}

	return channel.Response{Payload: []byte(address)}, nil
}
