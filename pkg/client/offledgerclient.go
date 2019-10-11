/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package client

import (
	"fmt"

	"github.com/bluele/gcache"
	"github.com/trustbloc/fabric-peer-ext/pkg/collections/client"
)

// OffLedger defines the functions of an off-ledger client
type OffLedger interface {
	Put(ns, coll, key string, value []byte) error
	Get(ns, coll, key string) ([]byte, error)
}

// OffLedgerProvider manages multiple off-ledger clients - one per channel
type OffLedgerProvider struct {
	cache gcache.Cache
}

// NewOffLedgerProvider returns a new off-ledger client provider
func NewOffLedgerProvider() *OffLedgerProvider {
	return &OffLedgerProvider{
		cache: gcache.New(0).LoaderFunc(func(channelID interface{}) (interface{}, error) {
			return client.New(channelID.(string)), nil
		}).Build(),
	}
}

// ForChannel returns the off-ledger client for the given channel
func (cp *OffLedgerProvider) ForChannel(channelID string) OffLedger {
	c, err := cp.cache.Get(channelID)
	if err != nil {
		// This should never happen since we never return an error in the loader func
		panic(fmt.Sprintf("unexpected error: %s", err))
	}
	return c.(OffLedger)
}
