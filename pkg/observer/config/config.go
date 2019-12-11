/*
Copyright SecureKey Technologies Inc. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package config

import (
	"time"
)

// Config stores observer configuration
type Config struct {
	peerID        string
	channels      []string
	monitorPeriod time.Duration
}

// New creates observer config
func New(peerID string, channels []string, monitorPeriod time.Duration) *Config {
	return &Config{
		peerID:        peerID,
		channels:      channels,
		monitorPeriod: monitorPeriod,
	}
}

// GetChannels returns the list of channels to observe for Sidetree transaction
func (cfg *Config) GetChannels() []string {
	return cfg.channels
}

// GetMonitorPeriod returns the period in which the monitor checks for presence of documents in the local DCAS store
func (cfg *Config) GetMonitorPeriod() time.Duration {
	return cfg.monitorPeriod
}

// GetPeerID returns the ID of the local peer
func (cfg *Config) GetPeerID() string {
	return cfg.peerID
}
