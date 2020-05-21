/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package config

import (
	"github.com/pkg/errors"

	"github.com/trustbloc/fabric-peer-ext/pkg/config/ledgerconfig/config"

	sidetreecfg "github.com/trustbloc/sidetree-fabric/pkg/config"
	"github.com/trustbloc/sidetree-fabric/pkg/rest/sidetreehandler"
)

// sidetreePeerValidator validates the SidetreePeer configuration
type sidetreePeerValidator struct {
	authTokenValidator *authTokenValidator
}

func newSidetreePeerValidator(provider tokenProvider) *sidetreePeerValidator {
	return &sidetreePeerValidator{
		authTokenValidator: newAuthTokenValidator(provider),
	}
}

func (v *sidetreePeerValidator) Validate(kv *config.KeyValue) error {
	if kv.AppName != SidetreePeerAppName {
		return nil
	}

	logger.Debugf("Validating config %s", kv)

	if kv.PeerID == "" {
		return errors.Errorf("field PeerID required for %s", kv.Key)
	}

	if kv.AppVersion != SidetreePeerAppVersion {
		return errors.Errorf("unsupported application version [%s] for %s", kv.AppVersion, kv.Key)
	}

	switch kv.ComponentName {
	case "":
		return v.validateConfig(kv)
	default:
		return v.validateHandlerConfig(kv)
	}
}

func (v *sidetreePeerValidator) validateConfig(kv *config.KeyValue) error {
	var sidetreeCfg sidetreecfg.SidetreePeer
	if err := unmarshal(kv.Value, &sidetreeCfg); err != nil {
		return errors.WithMessagef(err, "invalid config %s", kv.Key)
	}

	return v.validateObserver(kv, sidetreeCfg.Observer)
}

func (v *sidetreePeerValidator) validateHandlerConfig(kv *config.KeyValue) error {
	var cfg sidetreehandler.Config
	if err := unmarshal(kv.Value, &cfg); err != nil {
		return errors.WithMessagef(err, "invalid config %s", kv.Key)
	}

	if kv.ComponentVersion != SidetreeHandlerComponentVersion {
		return errors.Errorf("unsupported component version %s for %s", kv.ComponentVersion, kv.Key)
	}

	if cfg.BasePath == "" {
		return errors.Errorf("field 'BasePath' is required for %s", kv.Key)
	}

	if kv.ComponentName != cfg.BasePath {
		return errors.Errorf("invalid component name [%s] - component name must be set to the base path [%s] for %s", kv.ComponentName, cfg.BasePath, kv.Key)
	}

	if cfg.BasePath[0:1] != "/" {
		return errors.Errorf("field 'BasePath' must begin with '/' for %s", kv.Key)
	}

	if cfg.Namespace == "" {
		return errors.Errorf("field 'Namespace' is required for %s", kv.Key)
	}

	if err := v.authTokenValidator.Validate(cfg.Authorization, kv); err != nil {
		return err
	}

	return nil
}

func (v *sidetreePeerValidator) validateObserver(kv *config.KeyValue, cfg sidetreecfg.Observer) error {
	if cfg.Period == 0 {
		logger.Infof("The Sidetree observer period is set to 0 and therefore the default value will be used for [%s].", kv.PeerID)
	}

	if cfg.MaxAttempts == 0 {
		logger.Infof("Sidetree observer MaxAttempts is set to 0 and therefore the default value will be used for [%s].", kv.PeerID)
	}

	if cfg.MetaDataChaincodeName == "" {
		return errors.Errorf("field 'MetaDataChaincodeName' is required for %s", kv.Key)
	}

	return nil
}
