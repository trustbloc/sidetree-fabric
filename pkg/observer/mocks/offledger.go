/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"sync"

	"github.com/hyperledger/fabric-protos-go/ledger/queryresult"
	commonledger "github.com/hyperledger/fabric/common/ledger"
	"github.com/trustbloc/fabric-peer-ext/pkg/collections/client"
)

const (
	defaultQueryResultsKey = ""
)

// MockOffLedgerClient mocks the off-ledger client
type MockOffLedgerClient struct {
	sync.RWMutex
	m       map[string]map[string][]byte
	qr      map[string][]*queryresult.KV
	PutErr  error
	GetErr  error
	putErrs map[string]error
	getErrs map[string]error
}

// NewMockOffLedgerClient creates a mock off-ledger client
func NewMockOffLedgerClient() *MockOffLedgerClient {
	return &MockOffLedgerClient{
		m:       make(map[string]map[string][]byte),
		qr:      make(map[string][]*queryresult.KV),
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

// WithQueryResults sets the query results for the given query
func (m *MockOffLedgerClient) WithQueryResults(ns, coll, query string, results []*queryresult.KV) *MockOffLedgerClient {
	m.qr[getKey(ns, coll, query)] = results
	return m
}

// WithDefaultQueryResults sets the default query results that are returned if no result was found for a given query
func (m *MockOffLedgerClient) WithDefaultQueryResults(results []*queryresult.KV) *MockOffLedgerClient {
	m.qr[defaultQueryResultsKey] = results
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

// PutMultipleValues puts the given key/values
func (m *MockOffLedgerClient) PutMultipleValues(ns, coll string, kvs []*client.KeyValue) error {
	panic("not implemented")
}

// Delete deletes the given key(s)
func (m *MockOffLedgerClient) Delete(ns, coll string, keys ...string) error {
	panic("not implemented")
}

// GetMultipleKeys retrieves the values for the given keys
func (m *MockOffLedgerClient) GetMultipleKeys(ns, coll string, keys ...string) ([][]byte, error) {
	panic("not implemented")
}

// Query executes the given query and returns an iterator that contains results.
// Only used for state databases that support query.
// (Note that this function is not supported by transient data collections)
// The returned ResultsIterator contains results of type *KV which is defined in protos/ledger/queryresult.
func (m *MockOffLedgerClient) Query(ns, coll, query string) (commonledger.ResultsIterator, error) {
	qr, ok := m.qr[getKey(ns, coll, query)]
	if !ok {
		qr = m.qr[defaultQueryResultsKey]
	}
	return newResultsIterator(qr), nil
}

type resultsIterator struct {
	results []*queryresult.KV
	i       int
}

func newResultsIterator(results []*queryresult.KV) *resultsIterator {
	return &resultsIterator{
		results: results,
	}
}

func (it *resultsIterator) Next() (commonledger.QueryResult, error) {
	if it.i >= len(it.results) {
		return nil, nil
	}
	i := it.i
	it.i++
	return it.results[i], nil
}

func (it *resultsIterator) Close() {
}
