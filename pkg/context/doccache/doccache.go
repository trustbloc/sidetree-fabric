/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package doccache

import (
	"encoding/json"
	"fmt"

	"github.com/bluele/gcache"
	"github.com/hyperledger/fabric/common/flogging"
	"github.com/trustbloc/sidetree-core-go/pkg/dochandler"
	"github.com/trustbloc/sidetree-core-go/pkg/document"

	"github.com/trustbloc/sidetree-fabric/pkg/rest/sidetreehandler"
)

var logger = flogging.MustGetLogger("sidetree_context")

const defaultCacheSize = 10000

var errNotFound = fmt.Errorf("cache not found")

// Invalidator invalidates the given key in the document cache
type Invalidator interface {
	Invalidate(uniqueSuffix string)
}

// Provider manages document caches - one per channel/namespace combination
type Provider struct {
	cache gcache.Cache
}

// New returns a new document cache provider
func New() *Provider {
	return &Provider{
		cache: gcache.New(0).Build(),
	}
}

// CreateCachingOperationProcessor updates the document cache for the given channel ID and namespace using the given configuration. The previous
// cache is replaced with the new one.
func (p *Provider) CreateCachingOperationProcessor(channelID string, cfg sidetreehandler.Config, target dochandler.OperationProcessor) dochandler.OperationProcessor {
	logger.Infof("[%s:%s] Updating document cache", channelID, cfg.Namespace)

	cachingResolver := newCache(channelID, cfg, target)

	if err := p.cache.Set(
		cacheKey{channelID: channelID, namespace: cfg.Namespace},
		cachingResolver,
	); err != nil {
		// Should never return an error
		panic(err)
	}

	return cachingResolver
}

// GetDocumentInvalidator returns the invalidator for the given channel and namespace
func (p *Provider) GetDocumentInvalidator(channelID, namespace string) (Invalidator, error) {
	c, err := p.cache.Get(cacheKey{
		channelID: channelID,
		namespace: namespace,
	})
	if err != nil {
		if err == gcache.KeyNotFoundError {
			return nil, errNotFound
		}

		return nil, err
	}

	return c.(Invalidator), nil
}

type cacheKey struct {
	channelID, namespace string
}

type cache struct {
	channelID, namespace string
	target               dochandler.OperationProcessor
	cache                gcache.Cache
	marshal              func(v interface{}) ([]byte, error)
	unmarshal            func(data []byte, v interface{}) error
}

func newCache(channelID string, cfg sidetreehandler.Config, target dochandler.OperationProcessor) *cache {
	size := int(cfg.DocumentCacheSize)
	if size == 0 {
		size = defaultCacheSize
	}

	logger.Infof("[%s:%s] Creating document cache - Max Size: %d", channelID, cfg.Namespace, size)

	c := &cache{
		target:    target,
		channelID: channelID,
		namespace: cfg.Namespace,
		marshal:   json.Marshal,
		unmarshal: json.Unmarshal,
	}

	c.cache = gcache.New(size).ARC().LoaderFunc(func(key interface{}) (interface{}, error) {
		result, err := target.Resolve(key.(string))
		if err != nil {
			return nil, err
		}

		// Marshal the document into binary format so that every call to 'Resolve' returns a
		// copy of the document. This prevents the caller from modifying the cached document.
		resultBytes, err := c.marshal(result)
		if err != nil {
			return nil, err
		}

		logger.Debugf("[%s:%s] Caching document [%s]: %s", channelID, cfg.Namespace, key, resultBytes)

		return resultBytes, err
	}).Build()

	return c
}

// Resolve resolves the document for the given unique suffix
func (c *cache) Resolve(uniqueSuffix string) (*document.ResolutionResult, error) {
	v, err := c.cache.Get(uniqueSuffix)
	if err != nil {
		return nil, err
	}

	logger.Debugf("[%s:%s] Retrieved document from cache [%s]: %s", c.channelID, c.namespace, uniqueSuffix, v)

	r := &document.ResolutionResult{}
	err = c.unmarshal(v.([]byte), r)
	if err != nil {
		return nil, err
	}

	return r, nil
}

// Deletes removes the document for the given unique suffix
func (c *cache) Invalidate(uniqueSuffix string) {
	if c.cache.Remove(uniqueSuffix) {
		logger.Debugf("[%s:%s] Removed document for unique suffix [%s]", c.channelID, c.namespace, uniqueSuffix)
	} else {
		logger.Debugf("[%s:%s] Document not cached for unique suffix [%s]", c.channelID, c.namespace, uniqueSuffix)
	}
}
