/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"strings"
	"time"

	"github.com/hyperledger/fabric/peer/node"
	"github.com/spf13/viper"
	extpeer "github.com/trustbloc/fabric-peer-ext/pkg/peer"
	"github.com/trustbloc/fabric-peer-ext/pkg/resource"
	"github.com/trustbloc/sidetree-fabric/pkg/observer/config"
	sidetreepeer "github.com/trustbloc/sidetree-fabric/pkg/peer"
)

const (
	// Configure channels that will be observed for sidetree txn
	confObserverChannels = "ledger.sidetree.observer.channels"

	// confMonitorPeriod is the period in which the monitor checks for presence of documents in the local DCAS store
	confMonitorPeriod = "ledger.sidetree.monitor.period"

	// defaultMonitorPeriod is the default value for monitor period
	defaultMonitorPeriod = 5 * time.Second
)

func main() {
	setup()

	extpeer.Initialize()
	resource.Register(newConfigProvider)
	sidetreepeer.Initialize()

	if err := startPeer(); err != nil {
		panic(err)
	}
}

func setup() {
	// For environment variables.
	viper.SetEnvPrefix(node.CmdRoot)
	viper.AutomaticEnv()
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)

	node.InitCmd(nil, nil)
}

type configProvider struct {
	*config.Config
}

func newConfigProvider() *configProvider {
	peerID := viper.GetString("peer.id")
	if peerID == "" {
		panic("peer.id must be set")
	}

	return &configProvider{
		Config: config.New(
			peerID,
			getObserverChannels(),
			getMonitorPeriod()),
	}
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

func startPeer() error {
	return node.Start()
}
