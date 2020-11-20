/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package doccache

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/bluele/gcache"
	"github.com/hyperledger/fabric/common/flogging"
	"github.com/trustbloc/sidetree-core-go/pkg/dochandler"
	"github.com/trustbloc/sidetree-core-go/pkg/document"

	"github.com/trustbloc/sidetree-fabric/pkg/rest/sidetreehandler"
)

var logger = flogging.MustGetLogger("sidetree_context")

const defaultCacheSize = 10000

var errNotFound = fmt.Errorf("cache not found")

type gCache interface {
	Get(key interface{}) (interface{}, error)
	Has(key interface{}) bool
	Set(key, value interface{}) error
	SetWithExpire(interface{}, interface{}, time.Duration) error
}

// Invalidator invalidates the given key in the document cache
type Invalidator interface {
	Invalidate(uniqueSuffix string)
}

// Provider manages document caches - one per channel/namespace combination
type Provider struct {
	cache gCache
}

// New returns a new document cache provider
func New() *Provider {
	return &Provider{
		cache: gcache.New(0).Build(),
	}
}

// CreateCachingOperationProcessor updates the document cache for the given channel ID and namespace using the given configuration. The previous
// cache is replaced with the new one.
// The document cache stores the ResolutionResult in JSON format. Each document has an optional expiration which is specified in
// the Config.DocumentExpiry. If DocumentExpiry is 0 then the document never expires, but it still may be evicted to make
// room for other documents.
func (p *Provider) CreateCachingOperationProcessor(channelID string, cfg sidetreehandler.Config, target dochandler.OperationProcessor) dochandler.OperationProcessor {
	logger.Infof("[%s:%s] Updating document cache - DocumentExpiry: %s", channelID, cfg.Namespace, cfg.DocumentExpiry)

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

type cacheValue struct {
	stale       bool
	resultBytes []byte
}

type cache struct {
	sidetreehandler.Config
	channelID string
	target    dochandler.OperationProcessor
	cache     gCache
	marshal   func(v interface{}) ([]byte, error)
	unmarshal func(data []byte, v interface{}) error
}

func newCache(channelID string, cfg sidetreehandler.Config, target dochandler.OperationProcessor) *cache {
	size := int(cfg.DocumentCacheSize)
	if size == 0 {
		size = defaultCacheSize
	}

	logger.Infof("[%s:%s] Creating document cache - Max Size: %d", channelID, cfg.Namespace, size)

	c := &cache{
		Config:    cfg,
		target:    target,
		channelID: channelID,
		marshal:   json.Marshal,
		unmarshal: json.Unmarshal,
	}

	c.cache = gcache.New(size).ARC().LoaderExpireFunc(c.load).Build()

	return c
}

// Resolve resolves the document for the given unique suffix
func (c *cache) Resolve(uniqueSuffix string) (*document.ResolutionResult, error) {
	v, err := c.cache.Get(uniqueSuffix)
	if err != nil {
		return nil, err
	}

	cv := v.(*cacheValue)

	logger.Debugf("[%s:%s] Retrieved document from cache [%s]: Stale: %t, Value: %s", c.channelID, c.Namespace, uniqueSuffix, cv.stale, cv.resultBytes)

	var resultBytes []byte

	if cv.stale {
		// The cached document has been marked as stale. Check if a new document is available.
		resultBytes, err = c.loadAndUpdate(uniqueSuffix, cv.resultBytes)
		if err != nil {
			return nil, err
		}
	} else {
		resultBytes = cv.resultBytes
	}

	r := &document.ResolutionResult{}
	err = c.unmarshal(resultBytes, r)
	if err != nil {
		return nil, err
	}

	return r, nil
}

// Invalidate invalidates the entry for the given unique suffix. The entry is still cached but it is marked as stale so that subsequent
// requests know that the cached entry is stale (based on the hash of the value) and then retrieval is performed via the DB. Once a different
// entry is retrieved from the DB, the cache is updated with that entry.
func (c *cache) Invalidate(uniqueSuffix string) {
	if !c.cache.Has(uniqueSuffix) {
		logger.Debugf("[%s:%s] Document not cached for unique suffix [%s]", c.channelID, c.Namespace, uniqueSuffix)

		return
	}

	v, err := c.cache.Get(uniqueSuffix)
	if err != nil {
		logger.Errorf("[%s:%s] Error retrieving document from cache for unique suffix [%s]: err", c.channelID, c.Namespace, uniqueSuffix, err)

		return
	}

	cv := v.(*cacheValue)

	// Instead of deleting the document from the cache, mark it as stale
	cv.stale = true

	if c.DocumentExpiry > 0 {
		err = c.cache.SetWithExpire(uniqueSuffix, cv, c.DocumentExpiry)
	} else {
		err = c.cache.Set(uniqueSuffix, cv)
	}
	if err != nil {
		// Should never happen
		logger.Errorf("[%s:%s] Error caching document for unique suffix [%s]: %s", c.channelID, c.Namespace, uniqueSuffix, err)

		return
	}

	logger.Debugf("[%s:%s] Invalidated document for unique suffix [%s]", c.channelID, c.Namespace, uniqueSuffix)
}

func (c *cache) loadAndUpdate(uniqueSuffix string, staleBytes []byte) ([]byte, error) {
	logger.Debugf("[%s:%s] The cached document is stale [%s]. Retrieving from store...", c.channelID, c.Namespace, uniqueSuffix)

	value, exp, err := c.load(uniqueSuffix)
	if err != nil {
		logger.Debugf("[%s:%s] Error loading document [%s]: %s", c.channelID, c.Namespace, uniqueSuffix, err)

		return nil, err
	}

	cvLoaded := value.(*cacheValue)

	// If the loaded document is different from the cached document (which has already been marked as stale)
	// then update the cache with the new document
	if !bytes.Equal(cvLoaded.resultBytes, staleBytes) {
		logger.Infof("[%s:%s] The stale cached document has been updated. Caching new document [%s]: %s",
			c.channelID, c.Namespace, uniqueSuffix, cvLoaded.resultBytes)

		if exp != nil {
			err = c.cache.SetWithExpire(uniqueSuffix, cvLoaded, *exp)
		} else {
			err = c.cache.Set(uniqueSuffix, cvLoaded)
		}
		if err != nil {
			// Should never happen
			return nil, err
		}
	} else {
		logger.Infof("[%s:%s] Returning stale cached document since it has not yet been updated in the store [%s]: %s",
			c.channelID, c.Namespace, uniqueSuffix, cvLoaded.resultBytes)
	}

	return cvLoaded.resultBytes, nil
}

func (c *cache) load(uniqueSuffix interface{}) (interface{}, *time.Duration, error) {
	result, err := c.target.Resolve(uniqueSuffix.(string))
	if err != nil {
		return nil, nil, err
	}

	// Marshal the document into binary format so that every call to 'Resolve' returns a
	// copy of the document. This prevents the caller from modifying the cached document.
	resultBytes, err := c.marshal(result)
	if err != nil {
		return nil, nil, err
	}

	logger.Debugf("[%s:%s] Loaded document from store [%s]: %s", c.channelID, c.Namespace, uniqueSuffix, resultBytes)

	var exp *time.Duration
	if c.DocumentExpiry > 0 {
		exp = &c.DocumentExpiry
	}

	return &cacheValue{resultBytes: resultBytes}, exp, err
}
