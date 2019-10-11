/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"sync"
)

// MockOffLedgerClient mocks the off-ledger client
type MockOffLedgerClient struct {
	sync.RWMutex
	m       map[string]map[string][]byte
	PutErr  error
	GetErr  error
	putErrs map[string]error
	getErrs map[string]error
}

// NewMockOffLedgerClient creates a mock off-ledger client
func NewMockOffLedgerClient() *MockOffLedgerClient {
	return &MockOffLedgerClient{
		m:       make(map[string]map[string][]byte),
		putErrs: make(map[string]error),
		getErrs: make(map[string]error),
	}
}

// WithPutError injects an error when a call is made to put a value
func (m *MockOffLedgerClient) WithPutError(err error) *MockOffLedgerClient {
	m.PutErr = err
	return m
}

// WithGetError injects an error when a call is made to get a value
func (m *MockOffLedgerClient) WithGetError(err error) *MockOffLedgerClient {
	m.GetErr = err
	return m
}

// WithPutErrorForKey injects an error when a call is made to put a value for the given namespace, collection, and key
func (m *MockOffLedgerClient) WithPutErrorForKey(ns, coll, key string, err error) *MockOffLedgerClient {
	m.putErrs[getKey(ns, coll, key)] = err
	return m
}

// WithGetErrorForKey injects an error when a call is made to get a value for the given namespace, collection, and key
func (m *MockOffLedgerClient) WithGetErrorForKey(ns, coll, key string, err error) *MockOffLedgerClient {
	m.getErrs[getKey(ns, coll, key)] = err
	return m
}

func getKey(ns, coll, key string) string {
	return ns + "~" + coll + "~" + key
}

// Put stores content in the off-ledger store
func (m *MockOffLedgerClient) Put(ns, coll, key string, content []byte) error {
	if m.PutErr != nil {
		return m.PutErr
	}

	if err, ok := m.putErrs[getKey(ns, coll, key)]; ok {
		return err
	}

	m.Lock()
	defer m.Unlock()
	if _, ok := m.m[ns+coll]; !ok {
		m.m[ns+coll] = make(map[string][]byte)
	}

	m.m[ns+coll][key] = content

	return nil
}

// Get retrieves the value from mock content addressable storage
func (m *MockOffLedgerClient) Get(ns, coll, key string) ([]byte, error) {
	if m.GetErr != nil {
		return nil, m.GetErr
	}

	if err, ok := m.getErrs[getKey(ns, coll, key)]; ok {
		return nil, err
	}

	m.RLock()
	defer m.RUnlock()

	return m.m[ns+coll][key], nil
}

// GetMap retrieves all values for namespace and collection
func (m *MockOffLedgerClient) GetMap(ns, coll string) (map[string][]byte, error) {
	m.RLock()
	defer m.RUnlock()

	if _, ok := m.m[ns+coll]; !ok {
		return nil, nil
	}

	return m.m[ns+coll], nil
}
