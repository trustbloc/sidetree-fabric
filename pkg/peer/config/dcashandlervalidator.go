/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package config

import (
	"github.com/pkg/errors"

	"github.com/trustbloc/fabric-peer-ext/pkg/config/ledgerconfig/config"

	"github.com/trustbloc/sidetree-fabric/pkg/rest/dcashandler"
)

// dcasHandlerValidator validates the DCAS handler configuration
type dcasHandlerValidator struct {
}

func (v *dcasHandlerValidator) Validate(kv *config.KeyValue) error {
	if kv.AppName != DCASHandlerAppName {
		return nil
	}

	logger.Debugf("Validating DCAS handler config %s", kv)

	if kv.PeerID == "" {
		return errors.Errorf("field PeerID required for %s", kv.Key)
	}

	if kv.AppVersion != DCASHandlerAppVersion {
		return errors.Errorf("unsupported application version [%s] for %s", kv.AppVersion, kv.Key)
	}

	if kv.ComponentName == "" {
		return errors.Errorf("empty component name for %s", kv.Key)
	}

	if kv.ComponentVersion != DCASHandlerComponentVersion {
		return errors.Errorf("unsupported component version [%s] for %s", kv.ComponentVersion, kv.Key)
	}

	var cfg dcashandler.Config
	if err := unmarshal(kv.Value, &cfg); err != nil {
		return errors.WithMessagef(err, "invalid config %s", kv.Key)
	}

	logger.Debugf("Got DCAS handler config: %+v", cfg)

	if err := v.validateDCASHandler(cfg, kv); err != nil {
		return errors.WithMessagef(err, "error validating file handler for key %s", kv.Key)
	}

	if kv.ComponentName != cfg.BasePath {
		return errors.Errorf("invalid component name [%s] - component name must be set to the base path [%s] for %s", kv.ComponentName, cfg.BasePath, kv.Key)
	}

	return nil
}

func (v *dcasHandlerValidator) validateDCASHandler(cfg dcashandler.Config, kv *config.KeyValue) error {
	if cfg.BasePath == "" {
		return errors.Errorf("field 'BasePath' is required")
	}

	if cfg.BasePath[0:1] != "/" {
		return errors.Errorf("field 'BasePath' must begin with '/' for %s", kv.Key)
	}

	if cfg.ChaincodeName == "" {
		return errors.Errorf("field 'ChaincodeName' is required")
	}

	if cfg.Collection == "" {
		return errors.Errorf("field 'Collection' is required")
	}

	return nil
}
