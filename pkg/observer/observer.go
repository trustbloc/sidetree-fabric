/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package observer

import (
	"encoding/json"
	"time"

	"github.com/hyperledger/fabric/common/flogging"
	gossipapi "github.com/hyperledger/fabric/extensions/gossip/api"
	"github.com/hyperledger/fabric/extensions/gossip/blockpublisher"
	"github.com/pkg/errors"
	"github.com/trustbloc/fabric-peer-ext/pkg/roles"
	"github.com/trustbloc/sidetree-core-go/pkg/api/batch"
	sidetreeobserver "github.com/trustbloc/sidetree-core-go/pkg/observer"
	"github.com/trustbloc/sidetree-fabric/pkg/client"
	"github.com/trustbloc/sidetree-fabric/pkg/observer/common"
	"github.com/trustbloc/sidetree-fabric/pkg/observer/monitor"
	"github.com/trustbloc/sidetree-fabric/pkg/observer/notifier"
)

var logger = flogging.MustGetLogger("observer")

const (
	observerRole = "observer"
	sidetreeRole = "sidetree"
)

// publisher allows clients to add handlers for various block events
type publisher interface {
	// AddWriteHandler adds a handler for KV writes
	AddWriteHandler(handler gossipapi.WriteHandler)
}

// getBlockPublisher returns block publisher for channel
var getBlockPublisher = func(channelID string) publisher {
	return blockpublisher.GetProvider().ForChannel(channelID)
}

type cfg interface {
	GetChannels() []string
	GetMonitorPeriod() time.Duration
}

type dcasClientProvider interface {
	ForChannel(channelID string) client.DCAS
}

type offLedgerClientProvider interface {
	ForChannel(channelID string) client.OffLedger
}

type blockchainClientProvider interface {
	ForChannel(channelID string) (client.Blockchain, error)
}

type dcas struct {
	channelID      string
	clientProvider dcasClientProvider
}

func newDCAS(channelID string, provider dcasClientProvider) *dcas {
	return &dcas{
		channelID:      channelID,
		clientProvider: provider,
	}
}

func (d *dcas) Read(key string) ([]byte, error) {
	return d.getDCASClient().Get(common.SidetreeNs, common.SidetreeColl, key)
}

func (d *dcas) Put(ops []batch.Operation) error {
	for _, op := range ops {
		bytes, err := json.Marshal(op)
		if err != nil {
			return errors.Wrapf(err, "json marshal for op failed")
		}
		_, err = d.getDCASClient().Put(common.DocNs, common.DocColl, bytes)
		if err != nil {
			return errors.Wrapf(err, "dcas put failed")
		}
	}
	return nil
}

func (d *dcas) getDCASClient() client.DCAS {
	return d.clientProvider.ForChannel(d.channelID)
}

// Observer observes the ledger for new anchor files and updates the document store accordingly
type Observer struct {
	cfg          cfg
	docMonitor   *monitor.Monitor
	dcasProvider dcasClientProvider
}

// New returns a new Observer
func New(cfg cfg) *Observer {
	dcasProvider := getDCASClientProvider()
	docMonitor := monitor.New(
		&monitor.ClientProviders{
			Blockchain: getBlockchainClientProvider(),
			DCAS:       dcasProvider,
			OffLedger:  getOffLedgerClientProvider(),
		},
	)
	return &Observer{
		cfg:          cfg,
		dcasProvider: dcasProvider,
		docMonitor:   docMonitor,
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
		n := notifier.New(getBlockPublisher(channelID))
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

var getDCASClientProvider = func() dcasClientProvider {
	return client.NewDCASProvider()
}

var getOffLedgerClientProvider = func() offLedgerClientProvider {
	return client.NewOffLedgerProvider()
}

var getBlockchainClientProvider = func() blockchainClientProvider {
	return client.NewBlockchainProvider()
}
