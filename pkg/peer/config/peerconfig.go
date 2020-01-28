/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package config

import (
	"errors"
	"fmt"

	viper "github.com/spf13/viper2015"
)

const (
	sidetreeHostKey = "sidetree.host"
	sidetreePortKey = "sidetree.port"
)

// Peer holds the Sidetree peer config
type Peer struct {
	sidetreeHost string
	sidetreePort int
}

// NewPeer returns a new peer config
func NewPeer() *Peer {
	return &Peer{
		sidetreeHost: viper.GetString(sidetreeHostKey),
		sidetreePort: viper.GetInt(sidetreePortKey),
	}
}

// SidetreeListenURL returns the URL on which the Sidetree REST service should be started
func (c *Peer) SidetreeListenURL() (string, error) {
	host := c.sidetreeHost
	if host == "" {
		host = "0.0.0.0"
	}

	if c.sidetreePort == 0 {
		return "", errors.New("port is not set for REST service")
	}

	return fmt.Sprintf("%s:%d", host, c.sidetreePort), nil
}
