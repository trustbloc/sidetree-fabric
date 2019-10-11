/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package client

import (
	"fmt"

	"github.com/bluele/gcache"
	dcasclient "github.com/trustbloc/fabric-peer-ext/pkg/collections/offledger/dcas/client"
)

// DCAS defines the functions of a DCAS client
type DCAS interface {
	Get(ns, coll, key string) ([]byte, error)
	Put(ns, coll string, value []byte) (string, error)
}

// DCASProvider manages multiple DCAS clients - one per channel
type DCASProvider struct {
	cache gcache.Cache
}

// NewDCASProvider returns a new DCAS client provider
func NewDCASProvider() *DCASProvider {
	return &DCASProvider{
		cache: gcache.New(0).LoaderFunc(func(channelID interface{}) (interface{}, error) {
			return dcasclient.New(channelID.(string)), nil
		}).Build(),
	}
}

// ForChannel returns the DCAS client for the given channel
func (cp *DCASProvider) ForChannel(channelID string) DCAS {
	client, err := cp.cache.Get(channelID)
	if err != nil {
		// This should never happen since we never return an error in the loader func
		panic(fmt.Sprintf("unexpected error: %s", err))
	}
	return client.(DCAS)
}
