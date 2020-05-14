/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package sidetreesvc

import (
	gossipapi "github.com/hyperledger/fabric/extensions/gossip/api"

	"github.com/trustbloc/sidetree-fabric/pkg/config"
	"github.com/trustbloc/sidetree-fabric/pkg/context/common"
	"github.com/trustbloc/sidetree-fabric/pkg/observer"
	"github.com/trustbloc/sidetree-fabric/pkg/role"
)

type observerController struct {
	channelID string
	observer  *observer.Observer
}

func newObserverController(channelID string, peerConfig peerConfig, observerCfg config.Observer, dcasCfg config.DCAS, providers *observer.ClientProviders, opStoreProvider common.OperationStoreProvider, txnChan <-chan gossipapi.TxMetadata) *observerController {
	var o *observer.Observer
	if role.IsObserver() {
		o = observer.New(channelID, peerConfig, observerCfg, dcasCfg, providers, opStoreProvider, txnChan)
	}

	return &observerController{
		channelID: channelID,
		observer:  o,
	}
}

// Start starts the Sidetree observer if it is set
func (o *observerController) Start() error {
	if o.observer != nil {
		logger.Debugf("[%s] Starting Sidetree observer ...", o.channelID)

		return o.observer.Start()
	}

	return nil
}

// Stop stops the Sidetree observer if it is set
func (o *observerController) Stop() {
	if o.observer != nil {
		logger.Debugf("[%s] Stopping Sidetree observer ...", o.channelID)

		o.observer.Stop()
	}
}
