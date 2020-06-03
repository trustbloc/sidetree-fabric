/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package config

import (
	"github.com/pkg/errors"
	"github.com/trustbloc/sidetree-fabric/pkg/rest/discoveryhandler"

	"github.com/trustbloc/fabric-peer-ext/pkg/config/ledgerconfig/config"
)

// discoveryHandlerValidator validates the Discover handler configuration
type discoveryHandlerValidator struct {
	authTokenValidator *authTokenValidator
}

func newDiscoveryHandlerValidator(provider tokenProvider) *discoveryHandlerValidator {
	return &discoveryHandlerValidator{
		authTokenValidator: newAuthTokenValidator(provider),
	}
}

func (v *discoveryHandlerValidator) Validate(kv *config.KeyValue) error {
	if kv.AppName != DiscoveryHandlerAppName {
		return nil
	}

	logger.Debugf("Validating discovery handler config %s", kv)

	if kv.PeerID == "" {
		return errors.Errorf("field PeerID required for %s", kv.Key)
	}

	if kv.AppVersion != DiscoveryHandlerAppVersion {
		return errors.Errorf("unsupported application version [%s] for %s", kv.AppVersion, kv.Key)
	}

	if kv.ComponentName == "" {
		return errors.Errorf("empty component name for %s", kv.Key)
	}

	if kv.ComponentVersion != DiscoveryHandlerComponentVersion {
		return errors.Errorf("unsupported component version [%s] for %s", kv.ComponentVersion, kv.Key)
	}

	var cfg discoveryhandler.Config
	if err := unmarshal(kv.Value, &cfg); err != nil {
		return errors.WithMessagef(err, "invalid config %s", kv.Key)
	}

	logger.Debugf("Got discovery handler config: %+v", cfg)

	if err := v.validateDiscoveryHandler(cfg, kv); err != nil {
		return errors.WithMessagef(err, "error validating discovery handler for key %s", kv.Key)
	}

	if kv.ComponentName != cfg.BasePath {
		return errors.Errorf("invalid component name [%s] - component name must be set to the base path [%s] for %s", kv.ComponentName, cfg.BasePath, kv.Key)
	}

	return nil
}

func (v *discoveryHandlerValidator) validateDiscoveryHandler(cfg discoveryhandler.Config, kv *config.KeyValue) error {
	if cfg.BasePath == "" {
		return errors.Errorf("field 'BasePath' is required")
	}

	if cfg.BasePath[0:1] != "/" {
		return errors.Errorf("field 'BasePath' must begin with '/' for %s", kv.Key)
	}

	if err := v.authTokenValidator.Validate(cfg.Authorization, kv); err != nil {
		return err
	}

	return nil
}
