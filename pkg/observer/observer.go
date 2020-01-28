/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package observer

import (
	"encoding/json"

	"github.com/hyperledger/fabric/common/flogging"
	"github.com/pkg/errors"
	dcasclient "github.com/trustbloc/fabric-peer-ext/pkg/collections/offledger/dcas/client"
	"github.com/trustbloc/sidetree-core-go/pkg/api/batch"
	sidetreeobserver "github.com/trustbloc/sidetree-core-go/pkg/observer"
	"github.com/trustbloc/sidetree-fabric/pkg/observer/common"
	"github.com/trustbloc/sidetree-fabric/pkg/observer/notifier"
)

var logger = flogging.MustGetLogger("sidetree_observer")

type dcas struct {
	channelID      string
	clientProvider common.DCASClientProvider
}

func newDCAS(channelID string, provider common.DCASClientProvider) *dcas {
	return &dcas{
		channelID:      channelID,
		clientProvider: provider,
	}
}

func (d *dcas) Read(key string) ([]byte, error) {
	dcasClient, err := d.getDCASClient()
	if err != nil {
		return nil, err
	}
	return dcasClient.Get(common.SidetreeNs, common.SidetreeColl, key)
}

func (d *dcas) Put(ops []batch.Operation) error {
	for _, op := range ops {
		bytes, err := json.Marshal(op)
		if err != nil {
			return errors.Wrapf(err, "json marshal for op failed")
		}
		dcasClient, err := d.getDCASClient()
		if err != nil {
			return err
		}
		_, err = dcasClient.Put(common.DocNs, common.DocColl, bytes)
		if err != nil {
			return errors.Wrapf(err, "dcas put failed")
		}
	}
	return nil
}

func (d *dcas) getDCASClient() (dcasclient.DCAS, error) {
	return d.clientProvider.ForChannel(d.channelID)
}

// Observer observes the ledger for new anchor files and updates the document store accordingly
type Observer struct {
	channelID    string
	dcasProvider common.DCASClientProvider
	bpProvider   common.BlockPublisherProvider
}

// Providers are the providers required by the observer
type Providers struct {
	DCAS           common.DCASClientProvider
	OffLedger      common.OffLedgerClientProvider
	BlockPublisher common.BlockPublisherProvider
	Blockchain     common.BlockchainClientProvider
}

// New returns a new Observer
func New(channelID string, providers *Providers) *Observer {
	return &Observer{
		channelID:    channelID,
		dcasProvider: providers.DCAS,
		bpProvider:   providers.BlockPublisher,
	}
}

// Start starts channel observer
func (o *Observer) Start() error {
	logger.Infof("[%s] Starting observer for channel", o.channelID)

	// register to receive Sidetree transactions from blocks
	n := notifier.New(o.bpProvider.ForChannel(o.channelID))
	dcasVal := newDCAS(o.channelID, o.dcasProvider)
	sidetreeobserver.Start(n, dcasVal, dcasVal)

	return nil
}

// Stop stops the channel observer routines
func (o *Observer) Stop() {
	// TODO: Need to have a way of stopping the Observer
}
