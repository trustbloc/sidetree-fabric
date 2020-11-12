/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package doccache

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/trustbloc/sidetree-core-go/pkg/document"

	"github.com/trustbloc/sidetree-fabric/pkg/context/doccache/mocks"
	"github.com/trustbloc/sidetree-fabric/pkg/rest/sidetreehandler"
)

//go:generate counterfeiter -o ./mocks/operationprocessor.gen.go --fake-name OperationProcessor github.com/trustbloc/sidetree-core-go/pkg/dochandler.OperationProcessor

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

	result1 := &document.ResolutionResult{Document: doc1}
	result2 := &document.ResolutionResult{Document: doc2}

	resolver.ResolveReturnsOnCall(0, result1, nil)
	resolver.ResolveReturnsOnCall(1, nil, errNotFound)
	resolver.ResolveReturnsOnCall(2, result2, nil)

	c := newCache(channel1, cfg, resolver)
	require.NotNil(t, c)

	r, err := c.Resolve(docID1)
	require.NoError(t, err)
	require.NotNil(t, r)
	require.Equal(t, "key1", r.Document["key"])

	// Resolve with the same key again
	r, err = c.Resolve(docID1)
	require.NoError(t, err)
	require.NotNil(t, r)
	require.Equal(t, "key1", r.Document["key"])

	r, err = c.Resolve(docID2)
	require.EqualError(t, err, errNotFound.Error())
	require.Nil(t, r)

	c.Invalidate(docID1)
	c.Invalidate(docID2)

	r, err = c.Resolve(docID1)
	require.NoError(t, err)
	require.NotNil(t, r)
	require.Equal(t, "key2", r.Document["key"])
}

func TestDocumentCacheError(t *testing.T) {
	cfg := sidetreehandler.Config{Namespace: ns1}
	resolver := &mocks.OperationProcessor{}

	doc := make(document.Document)
	doc["key"] = "key1"

	resolver.ResolveReturns(&document.ResolutionResult{Document: doc}, nil)

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
}
