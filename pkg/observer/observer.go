/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package observer

import (
	"github.com/hyperledger/fabric/common/flogging"
	"github.com/pkg/errors"
	dcasclient "github.com/trustbloc/fabric-peer-ext/pkg/collections/offledger/dcas/client"
	sidetreeobserver "github.com/trustbloc/sidetree-core-go/pkg/observer"

	"github.com/trustbloc/sidetree-fabric/pkg/config"
	ctxcommon "github.com/trustbloc/sidetree-fabric/pkg/context/common"
	"github.com/trustbloc/sidetree-fabric/pkg/observer/common"
)

var logger = flogging.MustGetLogger("sidetree_observer")

type dcas struct {
	config.DCAS
	channelID      string
	clientProvider common.DCASClientProvider
}

func newDCAS(channelID string, dcasCfg config.DCAS, provider common.DCASClientProvider) *dcas {
	return &dcas{
		DCAS:           dcasCfg,
		channelID:      channelID,
		clientProvider: provider,
	}
}

func (d *dcas) Read(key string) ([]byte, error) {
	dcasClient, err := d.getDCASClient()
	if err != nil {
		return nil, err
	}

	data, err := dcasClient.Get(d.ChaincodeName, d.Collection, key)
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, errors.Errorf("content not found for key [%s]", key)
	}

	return data, nil
}

func (d *dcas) getDCASClient() (dcasclient.DCAS, error) {
	return d.clientProvider.ForChannel(d.channelID)
}

// Observer observes the ledger for new anchor files and updates the document store accordingly
type Observer struct {
	channelID string
	observer  *sidetreeobserver.Observer
}

// Providers are the providers required by the observer
type Providers struct {
	DCAS           common.DCASClientProvider
	OperationStore ctxcommon.OperationStoreProvider
	Ledger         sidetreeobserver.Ledger
	Filter         sidetreeobserver.OperationFilterProvider
}

// New returns a new Observer
func New(channelID string, dcasCfg config.DCAS, providers *Providers) *Observer {
	stProviders := &sidetreeobserver.Providers{
		Ledger:           providers.Ledger,
		DCASClient:       newDCAS(channelID, dcasCfg, providers.DCAS),
		OpStoreProvider:  storeProvider(providers.OperationStore),
		OpFilterProvider: providers.Filter,
	}

	return &Observer{
		channelID: channelID,
		observer:  sidetreeobserver.New(stProviders),
	}
}

// Start starts channel observer
func (o *Observer) Start() error {
	logger.Infof("[%s] Starting observer for channel", o.channelID)

	o.observer.Start()

	return nil
}

// Stop stops the channel observer routines
func (o *Observer) Stop() {
	logger.Infof("[%s] Stopping observer", o.channelID)

	o.observer.Stop()
}

type storePovider struct {
	opStoreProvider ctxcommon.OperationStoreProvider
}

func storeProvider(p ctxcommon.OperationStoreProvider) *storePovider {
	return &storePovider{opStoreProvider: p}
}

func (p *storePovider) ForNamespace(namespace string) (sidetreeobserver.OperationStore, error) {
	s, err := p.opStoreProvider.ForNamespace(namespace)
	if err != nil {
		return nil, err
	}

	return s, nil
}
