/*
Copyright SecureKey Technologies Inc. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"strings"

	"github.com/hyperledger/fabric/peer/node"
	"github.com/spf13/viper"
)

func main() {
	//!!!start other services before peer start

	// start peer
	if err := startPeer(); err != nil {
		panic(err.Error())
	}

}

func startPeer() error {
	// For environment variables.
	viper.SetEnvPrefix(node.CmdRoot)
	viper.AutomaticEnv()
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)

	node.InitCmd(nil, nil)

	return node.Start()
}
