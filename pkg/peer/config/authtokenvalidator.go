/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package config

import (
	"github.com/pkg/errors"
	"github.com/trustbloc/fabric-peer-ext/pkg/config/ledgerconfig/config"
	"github.com/trustbloc/sidetree-fabric/pkg/rest/authhandler"
)

type authTokenValidator struct {
	tokenProvider
}

func newAuthTokenValidator(provider tokenProvider) *authTokenValidator {
	return &authTokenValidator{tokenProvider: provider}
}

func (v *authTokenValidator) Validate(cfg authhandler.Config, kv *config.KeyValue) error {
	if len(cfg.ReadTokens) > 0 {
		if err := v.validate(cfg.ReadTokens, kv); err != nil {
			return err
		}
	} else {
		logger.Debugf("field 'ReadTokens' is not set for %s. No authorization will take place for reads on this endpoint.", kv.Key)
	}

	if len(cfg.WriteTokens) > 0 {
		if err := v.validate(cfg.WriteTokens, kv); err != nil {
			return err
		}
	} else {
		logger.Warnf("field 'WriteTokens' is not set for %s. No authorization will take place for writes to this endpoint.", kv.Key)
	}

	return nil
}

func (v *authTokenValidator) validate(tokens []string, kv *config.KeyValue) error {
	for _, token := range tokens {
		if v.SidetreeAPIToken(token) == "" {
			return errors.Errorf("token name [%s] is not defined in peer config for %s", token, kv.Key)
		}
	}

	return nil
}
