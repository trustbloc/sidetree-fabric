/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package observer

import (
	"encoding/json"
	"sync"

	"github.com/hyperledger/fabric/common/flogging"
	gossipapi "github.com/hyperledger/fabric/extensions/gossip/api"
	"github.com/hyperledger/fabric/extensions/gossip/blockpublisher"
	"github.com/pkg/errors"
	dcasclient "github.com/trustbloc/fabric-peer-ext/pkg/collections/offledger/dcas/client"
	"github.com/trustbloc/fabric-peer-ext/pkg/roles"
	"github.com/trustbloc/sidetree-core-go/pkg/api/batch"
	sidetreeobserver "github.com/trustbloc/sidetree-core-go/pkg/observer"
	"github.com/trustbloc/sidetree-fabric/pkg/observer/notifier"
)

var logger = flogging.MustGetLogger("observer")

const (
	sidetreeNs   = "sidetreetxn_cc"
	sidetreeColl = "dcas"
	docNs        = "document_cc"
	docColl      = "docs"
	observerRole = "observer"
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

type dcasClient interface {

	// Get retrieves the DCAS value
	Get(ns, coll, key string) ([]byte, error)

	// Put puts the DCAS value and returns the key for the value
	Put(ns, coll string, value []byte) (string, error)
}

var getDCAS = func(channelID string) dcasClient {
	return dcasclient.New(channelID)
}

type cfg interface {
	GetChannels() []string
}

type dcas struct {
	sync.RWMutex
	channelID string
	dcas      dcasClient
}

func (d *dcas) Read(key string) ([]byte, error) {
	return d.getDCASClient().Get(sidetreeNs, sidetreeColl, key)
}

func (d *dcas) Put(ops []batch.Operation) error {
	for _, op := range ops {
		bytes, err := json.Marshal(op)
		if err != nil {
			return errors.Wrapf(err, "json marshal for op failed")
		}
		_, err = d.getDCASClient().Put(docNs, docColl, bytes)
		if err != nil {
			return errors.Wrapf(err, "dcas put failed")
		}
	}
	return nil
}

func (d *dcas) getDCASClient() dcasClient {

	d.RLock()
	dcas := d.dcas
	d.RUnlock()

	if dcas != nil {
		return dcas
	}
	dcas = getDCAS(d.channelID)

	d.Lock()
	defer d.Unlock()

	d.dcas = dcas
	return dcas
}

// Start starts channel observer routines
func Start(cfg cfg) error {

	if roles.HasRole(observerRole) {

		logger.Infof("peer is an observer, channels to observe: %s", cfg.GetChannels())
		for _, channelID := range cfg.GetChannels() {
			// register to receive Sidetree transactions from blocks
			n := notifier.New(getBlockPublisher(channelID))
			dcasVal := &dcas{channelID: channelID}
			sidetreeobserver.Start(n, dcasVal, dcasVal)
		}

	} else {
		logger.Info("peer is not an observer, nothing to do...")
	}

	return nil
}
