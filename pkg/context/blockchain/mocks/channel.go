/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
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

// Execute mocks execute
func (cc *MockChannelClient) Execute(request channel.Request, options ...channel.RequestOption) (channel.Response, error) {
	if cc.Err != nil {
		return channel.Response{}, cc.Err
	}

	return channel.Response{}, nil
}
