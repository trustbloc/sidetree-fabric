/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package observer

import (
	"encoding/json"
	"time"

	"github.com/hyperledger/fabric/common/flogging"
	"github.com/pkg/errors"
	dcasclient "github.com/trustbloc/fabric-peer-ext/pkg/collections/offledger/dcas/client"
	"github.com/trustbloc/fabric-peer-ext/pkg/roles"
	"github.com/trustbloc/sidetree-core-go/pkg/api/batch"
	sidetreeobserver "github.com/trustbloc/sidetree-core-go/pkg/observer"
	"github.com/trustbloc/sidetree-fabric/pkg/observer/common"
	"github.com/trustbloc/sidetree-fabric/pkg/observer/monitor"
	"github.com/trustbloc/sidetree-fabric/pkg/observer/notifier"
)

var logger = flogging.MustGetLogger("observer")

const (
	observerRole = "observer"
	sidetreeRole = "sidetree"
)

type cfg interface {
	GetChannels() []string
	GetMonitorPeriod() time.Duration
}

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
	cfg          cfg
	docMonitor   *monitor.Monitor
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
func New(cfg cfg, providers *Providers) *Observer {
	docMonitor := monitor.New(
		&monitor.ClientProviders{
			Blockchain: providers.Blockchain,
			DCAS:       providers.DCAS,
			OffLedger:  providers.OffLedger,
		},
	)
	return &Observer{
		cfg:          cfg,
		dcasProvider: providers.DCAS,
		docMonitor:   docMonitor,
		bpProvider:   providers.BlockPublisher,
	}
}

// Start starts channel observer routines
func (o *Observer) Start() error {
	for _, channelID := range o.cfg.GetChannels() {
		if err := o.start(channelID); err != nil {
			return err
		}
	}
	return nil
}

// Stop stops the channel observer routines
func (o *Observer) Stop() {
	o.docMonitor.StopAll()
}

func (o *Observer) start(channelID string) error {
	if roles.HasRole(observerRole) {
		logger.Infof("Starting observer for channel [%s]", channelID)
		// register to receive Sidetree transactions from blocks
		n := notifier.New(o.bpProvider.ForChannel(channelID))
		dcasVal := newDCAS(channelID, o.dcasProvider)
		sidetreeobserver.Start(n, dcasVal, dcasVal)
		return nil
	}
	if roles.HasRole(sidetreeRole) && roles.IsCommitter() {
		return o.docMonitor.Start(channelID, o.cfg.GetMonitorPeriod())
	}
	logger.Debugf("Nothing to start for channel [%s]", channelID)
	return nil
}
