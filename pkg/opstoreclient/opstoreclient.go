/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package opstoreclient

import (
	"github.com/bluele/gcache"
	"github.com/trustbloc/sidetree-core-go/pkg/processor"
	"github.com/trustbloc/sidetree-fabric/pkg/context/store"
	"github.com/trustbloc/sidetree-fabric/pkg/observer/common"
)

// Provider manages operation store clients
type Provider struct {
	cache gcache.Cache
}

// NewProvider returns a new operation store client provider
func NewProvider(dcasProvider common.DCASClientProvider) *Provider {
	return &Provider{
		cache: gcache.New(0).LoaderFunc(func(key interface{}) (i interface{}, err error) {
			k := key.(providerKey)
			return store.New(k.channelID, k.namespace, dcasProvider), nil
		}).Build(),
	}
}

// Get returns the channel operation store client provider for the given channel and namespace
func (p *Provider) Get(channelID, namespace string) processor.OperationStoreClient {
	s, err := p.cache.Get(providerKey{channelID: channelID, namespace: namespace})
	if err != nil {
		// Shouldn't happen since the loader func never returns an error
		panic(err)
	}

	return s.(processor.OperationStoreClient)
}

type providerKey struct {
	channelID string
	namespace string
}
