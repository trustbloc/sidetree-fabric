/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package peer

import (
	"strings"
	"time"

	"github.com/hyperledger/fabric/common/flogging"
	"github.com/spf13/viper"
	"github.com/trustbloc/fabric-peer-ext/pkg/resource"
	"github.com/trustbloc/sidetree-fabric/pkg/client"
	"github.com/trustbloc/sidetree-fabric/pkg/observer"
	"github.com/trustbloc/sidetree-fabric/pkg/observer/config"
)

var logger = flogging.MustGetLogger("sidetree_peer")

const (
	// Configure channels that will be observed for sidetree txn
	confObserverChannels = "ledger.sidetree.observer.channels"

	// confMonitorPeriod is the period in which the monitor checks for presence of documents in the local DCAS store
	confMonitorPeriod = "ledger.sidetree.monitor.period"

	// defaultMonitorPeriod is the default value for monitor period
	defaultMonitorPeriod = 5 * time.Second
)

// Initialize initializes the required resources for peer startup
func Initialize() {
	resource.Register(client.NewBlockchainProvider)
	resource.Register(newObserver)
}

// getObserverChannels returns the channels that will be observed for Sidetree transaction
func getObserverChannels() []string {
	channels := viper.GetString(confObserverChannels)
	return strings.Split(channels, ",")
}

// getMonitorPeriod returns the period in which the monitor checks for presence of documents in the local DCAS store
func getMonitorPeriod() time.Duration {
	monitorPeriod := viper.GetDuration(confMonitorPeriod)
	if monitorPeriod == 0 {
		monitorPeriod = defaultMonitorPeriod
	}
	return monitorPeriod
}

type observerResource struct {
	observer *observer.Observer
}

func newObserver(providers *observer.Providers) *observerResource {
	logger.Infof("Initializing observer")

	observingChannels := getObserverChannels()
	monitorPeriod := getMonitorPeriod()
	cfg := config.New(observingChannels, monitorPeriod)

	o := observer.New(cfg, providers)
	if err := o.Start(); err != nil {
		panic(err)
	}

	return &observerResource{observer: o}
}

// Close stops the observer
func (r *observerResource) Close() {
	logger.Infof("Stopping observer")
	r.observer.Stop()
}
