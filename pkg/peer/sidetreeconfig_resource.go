/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package peer

import (
	"github.com/pkg/errors"
	viper "github.com/spf13/viper2015"
)

type sidetreeConfig struct {
	Channel string
}

func (c *sidetreeConfig) ChannelID() string {
	return c.Channel
}

func newSidetreeConfigResource() *sidetreeConfig {
	logger.Info("Creating Sidetree config")

	cfg, err := getSidetreeConfig()
	if err != nil {
		panic(err)
	}
	return cfg
}

func getSidetreeConfig() (*sidetreeConfig, error) {
	cfgFile := defaultConfigFile
	if viper.IsSet(keyConfigFile) {
		cfgFile = viper.GetString(keyConfigFile)
	}

	v := viper.New()
	v.SetConfigFile(cfgFile)
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	const key = "sidetree"

	if !v.IsSet(key) {
		return nil, errors.New("sidetree configuration key not found")
	}

	var sidetreeCfg sidetreeConfig
	if err := v.UnmarshalKey(key, &sidetreeCfg); err != nil {
		return nil, err
	}

	return &sidetreeCfg, nil
}
