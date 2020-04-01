/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package store

import (
	"github.com/bluele/gcache"

	commonledger "github.com/hyperledger/fabric/common/ledger"
	dcasclient "github.com/trustbloc/fabric-peer-ext/pkg/collections/offledger/dcas/client"

	"github.com/trustbloc/sidetree-fabric/pkg/config"
	"github.com/trustbloc/sidetree-fabric/pkg/context/common"
)

// Provider manages an operation store client for each namespace
type Provider struct {
	channelID    string
	cfgService   config.SidetreeService
	dcasProvider common.DCASClientProvider
	cache        gcache.Cache
}

// NewProvider returns a new operation store client provider
func NewProvider(channelID string, cfgServiceProvider config.SidetreeService, dcasProvider common.DCASClientProvider) *Provider {
	p := &Provider{
		channelID:    channelID,
		cfgService:   cfgServiceProvider,
		dcasProvider: dcasProvider,
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

	dcasClient, err := p.dcasProvider.ForChannel(p.channelID)
	if err != nil {
		return nil, err
	}

	return NewClient(
		p.channelID, namespace,
		newStore(dcasClient, cfg.ChaincodeName, cfg.Collection),
	), nil
}

type dataStore struct {
	dcasClient    dcasclient.DCAS
	chaincodeName string
	collection    string
}

func newStore(dcasClient dcasclient.DCAS, chaincodeName, collection string) *dataStore {
	return &dataStore{
		chaincodeName: chaincodeName,
		collection:    collection,
		dcasClient:    dcasClient,
	}
}

func (qp *dataStore) Query(query string) (commonledger.ResultsIterator, error) {
	return qp.dcasClient.Query(qp.chaincodeName, qp.collection, query)
}

func (qp *dataStore) Put(data []byte) error {
	_, err := qp.dcasClient.Put(qp.chaincodeName, qp.collection, data)
	return err
}
