/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package config

import (
	"github.com/pkg/errors"

	"github.com/trustbloc/fabric-peer-ext/pkg/config/ledgerconfig/config"

	"github.com/trustbloc/sidetree-fabric/pkg/rest/blockchainhandler"
)

// blockchainHandlerValidator validates the blockchain handler configuration
type blockchainHandlerValidator struct {
}

func (v *blockchainHandlerValidator) Validate(kv *config.KeyValue) error {
	if kv.AppName != BlockchainHandlerAppName {
		return nil
	}

	logger.Debugf("Validating blockchain handler config %s", kv)

	if kv.PeerID == "" {
		return errors.Errorf("field 'PeerID' required for %s", kv.Key)
	}

	if kv.AppVersion != BlockchainHandlerAppVersion {
		return errors.Errorf("unsupported application version [%s] for %s", kv.AppVersion, kv.Key)
	}

	if kv.ComponentName == "" {
		return errors.Errorf("empty component name for %s", kv.Key)
	}

	if kv.ComponentVersion != BlockchainHandlerComponentVersion {
		return errors.Errorf("unsupported component version [%s] for %s", kv.ComponentVersion, kv.Key)
	}

	var cfg blockchainhandler.Config
	if err := unmarshal(kv.Value, &cfg); err != nil {
		return errors.WithMessagef(err, "invalid config %s", kv.Key)
	}

	logger.Debugf("Got blockchain handler config: %+v", cfg)

	if err := v.validateBlockchainHandler(cfg, kv); err != nil {
		return errors.WithMessagef(err, "error validating blockchain handler for key %s", kv.Key)
	}

	if kv.ComponentName != cfg.BasePath {
		return errors.Errorf("invalid component name [%s] - component name must be set to the base path [%s] for %s", kv.ComponentName, cfg.BasePath, kv.Key)
	}

	return nil
}

func (v *blockchainHandlerValidator) validateBlockchainHandler(cfg blockchainhandler.Config, kv *config.KeyValue) error {
	if cfg.BasePath == "" {
		return errors.Errorf("field 'BasePath' is required")
	}

	if cfg.BasePath[0:1] != "/" {
		return errors.Errorf("field 'BasePath' must begin with '/' for %s", kv.Key)
	}

	if cfg.MaxTransactionsInResponse == 0 {
		logger.Warnf("field 'MaxTransactionsInResponse' is not set for %s. Will use default value.", kv.Key)
	}

	return nil
}
