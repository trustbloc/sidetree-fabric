/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"fmt"
	"sync"

	"github.com/trustbloc/fabric-peer-ext/pkg/collections/offledger/dcas"
)

// MockDCASClient mocks DCAS for testing purposes.
type MockDCASClient struct {
	sync.RWMutex
	m      map[string]map[string][]byte
	PutErr error
	GetErr error
}

// NewMockDCASClient creates mock DCAS client
func NewMockDCASClient() *MockDCASClient {
	return &MockDCASClient{m: make(map[string]map[string][]byte)}
}

// Put stores content in mock content addressable storage
func (m *MockDCASClient) Put(ns, coll string, content []byte) (string, error) {
	if m.PutErr != nil {
		return "", m.PutErr
	}

	m.Lock()
	defer m.Unlock()
	if _, ok := m.m[ns+coll]; !ok {
		m.m[ns+coll] = make(map[string][]byte)
	}

	key := dcas.GetCASKey(content)

	m.m[ns+coll][key] = content

	return key, nil

}

// Get retrieves the value from mock content addressable storage
func (m *MockDCASClient) Get(ns, coll, key string) ([]byte, error) {
	if m.GetErr != nil {
		return nil, m.GetErr
	}

	m.RLock()
	defer m.RUnlock()

	if _, ok := m.m[ns+coll]; !ok {
		return nil, fmt.Errorf("collection store doesn't exist")
	}

	value, ok := m.m[ns+coll][key]
	if !ok {
		return nil, fmt.Errorf("not found")
	}

	return value, nil
}

// GetMap retrieves all values for namespace and collection
func (m *MockDCASClient) GetMap(ns, coll string) (map[string][]byte, error) {

	m.RLock()
	defer m.RUnlock()

	if _, ok := m.m[ns+coll]; !ok {
		return nil, fmt.Errorf("collection store doesn't exist")
	}

	return m.m[ns+coll], nil
}
