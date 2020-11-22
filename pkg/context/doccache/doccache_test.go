/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package doccache

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/trustbloc/sidetree-core-go/pkg/api/protocol"
	"github.com/trustbloc/sidetree-core-go/pkg/document"

	"github.com/trustbloc/sidetree-fabric/pkg/context/doccache/mocks"
	"github.com/trustbloc/sidetree-fabric/pkg/rest/sidetreehandler"
)

//go:generate counterfeiter -o ./mocks/operationprocessor.gen.go --fake-name OperationProcessor github.com/trustbloc/sidetree-core-go/pkg/dochandler.OperationProcessor
//go:generate counterfeiter -o ./mocks/cache.gen.go --fake-name Cache . gCache

const (
	channel1 = "channel1"
	ns1      = "ns1"
	docID1   = "doc1"
	docID2   = "doc2"
)

func TestProvider(t *testing.T) {
	p := New()
	require.NotNil(t, p)

	c, err := p.GetDocumentInvalidator(channel1, ns1)
	require.EqualError(t, err, errNotFound.Error())
	require.Nil(t, c)

	cfg := sidetreehandler.Config{Namespace: ns1}
	resolver := &mocks.OperationProcessor{}

	processor := p.CreateCachingOperationProcessor(channel1, cfg, resolver)
	require.NotNil(t, processor)

	c, err = p.GetDocumentInvalidator(channel1, ns1)
	require.NoError(t, err)
	require.NotNil(t, c)
}

func TestDocumentCache(t *testing.T) {
	cfg := sidetreehandler.Config{Namespace: ns1}
	resolver := &mocks.OperationProcessor{}

	doc1 := make(document.Document)
	doc1["key"] = "key1"

	doc2 := make(document.Document)
	doc2["key"] = "key2"

	errNotFound := fmt.Errorf("not found")

	result1 := &protocol.ResolutionModel{Doc: doc1}
	result2 := &protocol.ResolutionModel{Doc: doc2}

	resolver.ResolveReturnsOnCall(0, result1, nil)
	resolver.ResolveReturnsOnCall(1, nil, errNotFound)
	resolver.ResolveReturnsOnCall(2, result1, nil)
	resolver.ResolveReturnsOnCall(3, result2, nil)

	c := newCache(channel1, cfg, resolver)
	require.NotNil(t, c)

	r, err := c.Resolve(docID1)
	require.NoError(t, err)
	require.NotNil(t, r)
	require.Equal(t, "key1", r.Doc["key"])

	// Resolve with the same key again
	r, err = c.Resolve(docID1)
	require.NoError(t, err)
	require.NotNil(t, r)
	require.Equal(t, "key1", r.Doc["key"])

	r, err = c.Resolve(docID2)
	require.EqualError(t, err, errNotFound.Error())
	require.Nil(t, r)

	c.Invalidate(docID1)
	c.Invalidate(docID2)

	// Resolve should return the stale result since it hasn't yet been updated in the store
	r, err = c.Resolve(docID1)
	require.NoError(t, err)
	require.NotNil(t, r)
	require.Equal(t, "key1", r.Doc["key"])

	// Resolve should return the new result
	r, err = c.Resolve(docID1)
	require.NoError(t, err)
	require.NotNil(t, r)
	require.Equal(t, "key2", r.Doc["key"])

	time.Sleep(300 * time.Millisecond)

	r, err = c.Resolve(docID1)
	require.NoError(t, err)
	require.NotNil(t, r)
	require.Equal(t, "key2", r.Doc["key"])

	require.Equal(t, 4, resolver.ResolveCallCount())
}

func TestDocumentCacheWithExpiry(t *testing.T) {
	cfg := sidetreehandler.Config{Namespace: ns1, DocumentExpiry: 200 * time.Millisecond}
	resolver := &mocks.OperationProcessor{}

	doc1 := make(document.Document)
	doc1["key"] = "key1"

	doc2 := make(document.Document)
	doc2["key"] = "key2"

	errNotFound := fmt.Errorf("not found")

	result1 := &protocol.ResolutionModel{Doc: doc1}
	result2 := &protocol.ResolutionModel{Doc: doc2}

	resolver.ResolveReturnsOnCall(0, result1, nil)
	resolver.ResolveReturnsOnCall(1, nil, errNotFound)
	resolver.ResolveReturnsOnCall(2, result2, nil)
	resolver.ResolveReturnsOnCall(3, result2, nil)

	c := newCache(channel1, cfg, resolver)
	require.NotNil(t, c)

	r, err := c.Resolve(docID1)
	require.NoError(t, err)
	require.NotNil(t, r)
	require.Equal(t, "key1", r.Doc["key"])

	// Resolve with the same key again
	r, err = c.Resolve(docID1)
	require.NoError(t, err)
	require.NotNil(t, r)
	require.Equal(t, "key1", r.Doc["key"])

	r, err = c.Resolve(docID2)
	require.EqualError(t, err, errNotFound.Error())
	require.Nil(t, r)

	c.Invalidate(docID1)
	c.Invalidate(docID2)

	r, err = c.Resolve(docID1)
	require.NoError(t, err)
	require.NotNil(t, r)
	require.Equal(t, "key2", r.Doc["key"])

	time.Sleep(300 * time.Millisecond)

	r, err = c.Resolve(docID1)
	require.NoError(t, err)
	require.NotNil(t, r)
	require.Equal(t, "key2", r.Doc["key"])

	require.Equal(t, 4, resolver.ResolveCallCount())
}

func TestDocumentCacheError(t *testing.T) {
	cfg := sidetreehandler.Config{Namespace: ns1}
	resolver := &mocks.OperationProcessor{}

	doc := make(document.Document)
	doc["key"] = "key1"

	resolver.ResolveReturns(&protocol.ResolutionModel{Doc: doc}, nil)

	c := newCache(channel1, cfg, resolver)
	require.NotNil(t, c)

	t.Run("marshal error", func(t *testing.T) {
		errExpected := fmt.Errorf("marshal error")
		c.marshal = func(interface{}) ([]byte, error) { return nil, errExpected }
		defer func() { c.marshal = json.Marshal }()

		r, err := c.Resolve(docID1)
		require.EqualError(t, err, errExpected.Error())
		require.Nil(t, r)
	})

	t.Run("marshal error", func(t *testing.T) {
		errExpected := fmt.Errorf("unmarshal error")
		c.unmarshal = func([]byte, interface{}) error { return errExpected }
		defer func() { c.unmarshal = json.Unmarshal }()

		r, err := c.Resolve(docID1)
		require.EqualError(t, err, errExpected.Error())
		require.Nil(t, r)
	})

	t.Run("Cache Set error", func(t *testing.T) {
		errExpected := fmt.Errorf("injected Set error")
		cache := &mocks.Cache{}
		cache.GetReturns(&cacheValue{stale: true}, nil)
		cache.SetReturns(errExpected)

		restoreCache := c.cache
		c.cache = cache
		defer func() { c.cache = restoreCache }()

		r, err := c.Resolve(docID1)
		require.EqualError(t, err, errExpected.Error())
		require.Nil(t, r)
	})

	t.Run("Cache Set error on Invalidate", func(t *testing.T) {
		errExpected := fmt.Errorf("injected Set error")
		cache := &mocks.Cache{}
		cache.HasReturns(true)
		cache.GetReturns(&cacheValue{stale: true}, nil)
		cache.SetReturns(errExpected)

		restoreCache := c.cache
		c.cache = cache
		defer func() { c.cache = restoreCache }()

		require.NotPanics(t, func() { c.Invalidate(docID1) })
	})

	t.Run("Cache Get error on Invalidate", func(t *testing.T) {
		errExpected := fmt.Errorf("injected Get error")
		cache := &mocks.Cache{}
		cache.HasReturns(true)
		cache.GetReturns(&cacheValue{stale: true}, errExpected)

		restoreCache := c.cache
		c.cache = cache
		defer func() { c.cache = restoreCache }()

		require.NotPanics(t, func() { c.Invalidate(docID1) })
	})

	t.Run("Cache Get error on loadAndUpdate", func(t *testing.T) {
		errExpected := fmt.Errorf("injected Get error")

		resolver.ResolveReturns(nil, errExpected)
		defer resolver.ResolveReturns(&protocol.ResolutionModel{Doc: doc}, nil)

		cache := &mocks.Cache{}
		cache.HasReturns(true)
		cache.GetReturnsOnCall(0, &cacheValue{stale: true}, nil)
		cache.GetReturnsOnCall(1, nil, errExpected)

		restoreCache := c.cache
		c.cache = cache
		defer func() { c.cache = restoreCache }()

		r, err := c.Resolve(docID1)
		require.EqualError(t, err, errExpected.Error())
		require.Nil(t, r)
	})
}
