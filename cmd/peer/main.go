/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"strings"

	"github.com/hyperledger/fabric/peer/node"
	viper "github.com/spf13/viper2015"
	extpeer "github.com/trustbloc/fabric-peer-ext/pkg/peer"
	sidetreepeer "github.com/trustbloc/sidetree-fabric/pkg/peer"
)

func main() {
	setup()

	extpeer.Initialize()
	sidetreepeer.Initialize()

	if err := startPeer(); err != nil {
		panic(err)
	}
}

func setup() {
	replacer := strings.NewReplacer(".", "_")

	viper.SetEnvPrefix(node.CmdRoot)
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(replacer)

	node.InitCmd(nil, nil)
}

func startPeer() error {
	return node.Start()
}
