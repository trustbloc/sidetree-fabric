/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package store

import (
	"crypto"
	"hash"

	"github.com/bluele/gcache"
	commonledger "github.com/hyperledger/fabric/common/ledger"
	mh "github.com/multiformats/go-multihash"
	"github.com/trustbloc/fabric-peer-ext/pkg/collections/client"

	"github.com/trustbloc/sidetree-fabric/pkg/config"
	"github.com/trustbloc/sidetree-fabric/pkg/context/common"
)

// Provider manages an operation store client for each namespace
type Provider struct {
	channelID         string
	cfgService        config.SidetreeService
	offLedgerProvider common.OffLedgerClientProvider
	cache             gcache.Cache
}

// NewProvider returns a new operation store client provider
func NewProvider(channelID string, cfgServiceProvider config.SidetreeService, offLedgerProvider common.OffLedgerClientProvider) *Provider {
	p := &Provider{
		channelID:         channelID,
		cfgService:        cfgServiceProvider,
		offLedgerProvider: offLedgerProvider,
	}

	p.cache = gcache.New(0).LoaderFunc(func(namespace interface{}) (i interface{}, err error) {
		return p.createClient(namespace.(string))
	}).Build()

	return p
}

// ForNamespace returns the operation store client for the given namespace
func (p *Provider) ForNamespace(namespace string) (common.OperationStore, error) {
	s, err := p.cache.Get(namespace)
	if err != nil {
		return nil, err
	}

	return s.(*Client), nil
}

func (p *Provider) createClient(namespace string) (*Client, error) {
	cfg, err := p.cfgService.LoadSidetree(namespace)
	if err != nil {
		return nil, err
	}

	offLedgerClient, err := p.offLedgerProvider.ForChannel(p.channelID)
	if err != nil {
		return nil, err
	}

	return NewClient(
		p.channelID, namespace,
		newStore(offLedgerClient, cfg.ChaincodeName, cfg.Collection),
	), nil
}

type dataStore struct {
	offLedgerClient client.OffLedger
	chaincodeName   string
	collection      string
	getHash         func() hash.Hash
}

func newStore(offLedgerClient client.OffLedger, chaincodeName, collection string) *dataStore {
	return &dataStore{
		chaincodeName:   chaincodeName,
		collection:      collection,
		offLedgerClient: offLedgerClient,
		getHash:         crypto.SHA256.New,
	}
}

func (qp *dataStore) Query(query string) (commonledger.ResultsIterator, error) {
	return qp.offLedgerClient.Query(qp.chaincodeName, qp.collection, query)
}

func (qp *dataStore) Put(data []byte) error {
	// Generate a base58 encoded key from the content. The key doesn't really matter but it must be unique.
	key, err := qp.getKey(data)
	if err != nil {
		return err
	}

	return qp.offLedgerClient.Put(qp.chaincodeName, qp.collection, key, data)
}

// getKey returns a base58-encoded key for the given bytes
func (qp *dataStore) getKey(data []byte) (string, error) {
	h := qp.getHash()

	_, err := h.Write(data)
	if err != nil {
		return "", err
	}

	return mh.Multihash(h.Sum(nil)).B58String(), nil
}
