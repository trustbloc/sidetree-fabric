/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"github.com/trustbloc/fabric-peer-ext/pkg/collections/offledger/dcas"
)

// MockDCASClient mocks DCAS for testing purposes.
type MockDCASClient struct {
	*MockOffLedgerClient
}

// NewMockDCASClient creates mock DCAS client
func NewMockDCASClient() *MockDCASClient {
	return &MockDCASClient{
		MockOffLedgerClient: NewMockOffLedgerClient(),
	}
}

// Put stores content in mock content addressable storage
func (m *MockDCASClient) Put(ns, coll string, content []byte) (string, error) {
	key, value, err := dcas.GetCASKeyAndValue(content)
	if err != nil {
		return "", err
	}
	return key, m.MockOffLedgerClient.Put(ns, coll, key, value)
}

// PutMultipleValues puts the DCAS values and returns the keys for the values
func (m *MockDCASClient) PutMultipleValues(ns, coll string, values [][]byte) ([]string, error) {
	panic("not implemented")
}
