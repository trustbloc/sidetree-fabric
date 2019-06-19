/*
Copyright SecureKey Technologies Inc. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package config

// Config stores observer configuration
type Config struct {
	channels []string
}

// New creates observer config
func New(channels []string) *Config {
	return &Config{channels: channels}
}

// GetChannels returns the list of channels to observe for Sidetree transaction
func (cfg *Config) GetChannels() []string {
	return cfg.channels
}
