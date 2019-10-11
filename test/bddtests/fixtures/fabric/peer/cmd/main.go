/*
Copyright SecureKey Technologies Inc. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"strings"

	"github.com/trustbloc/sidetree-fabric/pkg/observer"
	"github.com/trustbloc/sidetree-fabric/pkg/observer/config"

	"github.com/hyperledger/fabric/peer/node"
	"github.com/spf13/viper"
)

// Configure channels that will be observed for sidetree txn
const confObserverChannels = "ledger.sidetree.observer.channels"

func main() {

	setup()

	// start additional services here before starting peer
	if err := startObserver(); err != nil {
		panic(err)
	}

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

func startObserver() error {
	observingChannels := getObserverChannels()
	cfg := config.New(observingChannels)
	return observer.New(cfg).Start()
}

// getObserverChannels returns the channels that will be observed for Sidetree transaction
func getObserverChannels() []string {
	channels := viper.GetString(confObserverChannels)
	return strings.Split(channels, ",")
}

func startPeer() error {
	return node.Start()
}
