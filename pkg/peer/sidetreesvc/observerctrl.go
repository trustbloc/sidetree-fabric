/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package sidetreesvc

import (
	sidetreeobserver "github.com/trustbloc/sidetree-core-go/pkg/observer"

	"github.com/trustbloc/sidetree-fabric/pkg/context/common"
	"github.com/trustbloc/sidetree-fabric/pkg/observer"
	"github.com/trustbloc/sidetree-fabric/pkg/observer/operationfilter"
	"github.com/trustbloc/sidetree-fabric/pkg/role"
)

type observerController struct {
	channelID string
	observer  *observer.Observer
}

func newObserverController(channelID string, dcasProvider common.DCASClientProvider, opStoreProvider common.OperationStoreProvider, notifier sidetreeobserver.Ledger) *observerController {
	var o *observer.Observer

	if role.IsObserver() {
		o = observer.New(channelID,
			&observer.Providers{
				DCAS:           dcasProvider,
				OperationStore: opStoreProvider,
				Ledger:         notifier,
				Filter:         operationfilter.NewProvider(channelID, opStoreProvider),
			},
		)
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
