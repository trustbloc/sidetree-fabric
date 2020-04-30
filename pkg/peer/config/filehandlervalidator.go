/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package config

import (
	"strings"

	"github.com/pkg/errors"

	"github.com/trustbloc/fabric-peer-ext/pkg/config/ledgerconfig/config"

	"github.com/trustbloc/sidetree-fabric/pkg/rest/filehandler"
)

// fileHandlerValidator validates the file handler configuration
type fileHandlerValidator struct {
	authTokenValidator *authTokenValidator
}

func newFileHandlerValidator(provider tokenProvider) *fileHandlerValidator {
	return &fileHandlerValidator{
		authTokenValidator: newAuthTokenValidator(provider),
	}
}

func (v *fileHandlerValidator) Validate(kv *config.KeyValue) error {
	if kv.AppName != FileHandlerAppName {
		return nil
	}

	logger.Debugf("Validating file handler config %s", kv)

	if kv.PeerID == "" {
		return errors.Errorf("field PeerID required for %s", kv.Key)
	}

	if kv.AppVersion != FileHandlerAppVersion {
		return errors.Errorf("unsupported application version [%s] for %s", kv.AppVersion, kv.Key)
	}

	if kv.ComponentName == "" {
		return errors.Errorf("empty component name for %s", kv.Key)
	}

	var cfg filehandler.Config
	if err := unmarshal(kv.Value, &cfg); err != nil {
		return errors.WithMessagef(err, "invalid config %s", kv.Key)
	}

	logger.Debugf("Got file handler config: %+v", cfg)

	if err := v.validateFileHandler(cfg, kv); err != nil {
		return errors.WithMessagef(err, "error validating file handler for key %s", kv.Key)
	}

	if kv.ComponentName != cfg.BasePath {
		return errors.Errorf("invalid component name [%s] - component name must be set to the base path [%s] for %s", kv.ComponentName, cfg.BasePath, kv.Key)
	}

	return nil
}

func (v *fileHandlerValidator) validateFileHandler(cfg filehandler.Config, kv *config.KeyValue) error {
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

	if cfg.IndexNamespace == "" {
		return errors.Errorf("field 'IndexNamespace' is required")
	}

	if cfg.IndexDocID == "" {
		logger.Warnf("Warning for key [%s]: Field 'IndexDocID' was not provided for [%s]. Files can be uploaded but they cannot be retrieved until a valid 'IndexDocID' is provided.", kv.Key, cfg.BasePath)
	} else if !strings.HasPrefix(cfg.IndexDocID, cfg.IndexNamespace) {
		return errors.Errorf("field 'IndexDocID' must begin with '%s'", cfg.IndexNamespace)
	}

	if err := v.authTokenValidator.Validate(cfg.Authorization, kv); err != nil {
		return err
	}

	return nil
}
