/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"sort"
	"strings"

	"github.com/hyperledger/fabric/core/ledger/kvledger/txmgmt/statedb"
)

// StateDB implements a mock StateDB
type StateDB struct {
	state             map[string]map[string][]byte
	queryResults      map[string][]*statedb.VersionedKV
	error             error
	queryError        error
	itProvider        func() *KVIterator
	bytesKeySupported bool
}

// NewStateDB returns a new mock state DB
func NewStateDB() *StateDB {
	return &StateDB{
		state:        make(map[string]map[string][]byte),
		queryResults: make(map[string][]*statedb.VersionedKV),
		itProvider:   NewKVIterator,
	}
}

// WithState sets the state
func (m *StateDB) WithState(ns, key string, value []byte) *StateDB {
	nsState, ok := m.state[ns]
	if !ok {
		nsState = make(map[string][]byte)
		m.state[ns] = nsState
	}
	nsState[key] = value
	return m
}

// WithPrivateState sets the private state
func (m *StateDB) WithPrivateState(ns, collection, key string, value []byte) *StateDB {
	nskey := privateNamespace(ns, collection)
	nsState, ok := m.state[nskey]
	if !ok {
		nsState = make(map[string][]byte)
		m.state[nskey] = nsState
	}
	nsState[key] = value
	return m
}

// WithDeletedState deletes the state
func (m *StateDB) WithDeletedState(ns, key string) *StateDB {
	nsState, ok := m.state[ns]
	if ok {
		delete(nsState, key)
	}
	return m
}

// WithDeletedPrivateState deletes the state
func (m *StateDB) WithDeletedPrivateState(ns, collection, key string) *StateDB {
	nsState, ok := m.state[privateNamespace(ns, collection)]
	if ok {
		delete(nsState, key)
	}
	return m
}

// WithQueryResults sets the query results for a given query on a namespace
func (m *StateDB) WithQueryResults(ns, query string, results []*statedb.VersionedKV) *StateDB {
	m.queryResults[queryResultsKey(ns, query)] = results
	return m
}

// WithPrivateQueryResults sets the query results for a given query on a private collection
func (m *StateDB) WithPrivateQueryResults(ns, coll, query string, results []*statedb.VersionedKV) *StateDB {
	m.queryResults[privateQueryResultsKey(ns, coll, query)] = results
	return m
}

// WithIteratorProvider sets the results iterator provider
func (m *StateDB) WithIteratorProvider(p func() *KVIterator) *StateDB {
	m.itProvider = p
	return m
}

// WithError injects an error to the mock executor
func (m *StateDB) WithError(err error) *StateDB {
	m.error = err
	return m
}

// WithQueryError injects an error to the mock executor for queries
func (m *StateDB) WithQueryError(err error) *StateDB {
	m.queryError = err
	return m
}

// GetState gets the value for given namespace and key. For a chaincode, the namespace corresponds to the chaincodeId
func (m *StateDB) GetState(namespace string, key string) ([]byte, error) {
	if m.error != nil {
		return nil, m.error
	}

	return m.state[namespace][key], nil
}

// GetStateMultipleKeys gets the values for multiple keys in a single call
func (m *StateDB) GetStateMultipleKeys(namespace string, keys []string) ([][]byte, error) {
	values := make([][]byte, len(keys))
	for i, k := range keys {
		v, err := m.GetState(namespace, k)
		if err != nil {
			return nil, err
		}
		values[i] = v
	}
	return values, nil
}

// GetStateRangeScanIterator returns an iterator that contains all the key-values between given key ranges.
// startKey is inclusive
// endKey is exclusive
// The returned ResultsIterator contains results of type *VersionedKV
func (m *StateDB) GetStateRangeScanIterator(namespace string, startKey string, endKey string) (statedb.ResultsIterator, error) {
	if m.queryError != nil {
		return nil, m.queryError
	}

	var kvs []*statedb.VersionedKV
	for key, value := range m.state[namespace] {
		if strings.Compare(key, startKey) < 0 || strings.Compare(key, endKey) >= 0 {
			continue
		}
		kvs = append(kvs, &statedb.VersionedKV{
			CompositeKey: statedb.CompositeKey{
				Namespace: namespace,
				Key:       key,
			},
			VersionedValue: statedb.VersionedValue{
				Value: value,
			},
		})
	}
	sort.Slice(kvs, func(i, j int) bool {
		return strings.Compare(kvs[i].Key, kvs[j].Key) < 0
	})

	return m.itProvider().WithResults(kvs...), nil
}

// GetStateRangeScanIteratorWithPagination returns an iterator that contains all the key-values between given key ranges.
// startKey is inclusive
// endKey is exclusive
// pageSize parameter limits the number of returned results
// The returned ResultsIterator contains results of type *VersionedKV
func (m *StateDB) GetStateRangeScanIteratorWithPagination(namespace string, startKey string, endKey string, pageSize int32) (statedb.QueryResultsIterator, error) {
	panic("not implemented")
}

// ExecuteQuery executes the given query and returns an iterator that contains results of type *VersionedKV.
func (m *StateDB) ExecuteQuery(namespace, query string) (statedb.ResultsIterator, error) {
	if m.queryError != nil {
		return nil, m.queryError
	}
	return m.itProvider().WithResults(m.queryResults[queryResultsKey(namespace, query)]...), m.error
}

// ExecuteQueryWithPagination executes the given query and
// returns an iterator that contains results of type *VersionedKV.
// The bookmark and page size parameters are associated with the pagination query.
func (m *StateDB) ExecuteQueryWithPagination(namespace, query, bookmark string, pageSize int32) (statedb.QueryResultsIterator, error) {
	panic("not implemented")
}

// BytesKeySupported returns true if the implementation (underlying db) supports the any bytes to be used as key.
// For instance, leveldb supports any bytes for the key while the couchdb supports only valid utf-8 string
func (m *StateDB) BytesKeySupported() bool {
	return m.bytesKeySupported
}

func privateNamespace(namespace, collection string) string {
	return namespace + "$" + collection
}

func queryResultsKey(namespace, query string) string {
	return namespace + "~" + query
}

func privateQueryResultsKey(namespace, coll, query string) string {
	return privateNamespace(namespace, coll) + "~" + query
}

// KVIterator is a mock key-value iterator
type KVIterator struct {
	kvs     []*statedb.VersionedKV
	nextIdx int
	err     error
}

// NewKVIterator returns a mock key-value iterator
func NewKVIterator() *KVIterator {
	return &KVIterator{}
}

// WithResults adds a key-value
func (it *KVIterator) WithResults(kv ...*statedb.VersionedKV) *KVIterator {
	it.kvs = append(it.kvs, kv...)
	return it
}

// WithError injects an error
func (it *KVIterator) WithError(err error) *KVIterator {
	it.err = err
	return it
}

// Next returns the next item in the result set. The `QueryResult` is expected to be nil when
// the iterator gets exhausted
func (it *KVIterator) Next() (statedb.QueryResult, error) {
	if it.err != nil {
		return nil, it.err
	}

	if it.nextIdx >= len(it.kvs) {
		return nil, nil
	}
	qr := it.kvs[it.nextIdx]
	it.nextIdx++
	return qr, nil
}

// GetBookmarkAndClose is not implemented
func (it *KVIterator) GetBookmarkAndClose() string {
	panic("not implemented")
}

// Close releases resources occupied by the iterator
func (it *KVIterator) Close() {
}
