/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package config

import (
	"github.com/pkg/errors"

	"github.com/trustbloc/fabric-peer-ext/pkg/config/ledgerconfig/config"

	sidetreecfg "github.com/trustbloc/sidetree-fabric/pkg/config"
)

// dcasValidator validates the Sidetree configuration including Protocols
type dcasValidator struct {
}

func (v *dcasValidator) Validate(kv *config.KeyValue) error {
	if kv.AppName != DCASAppName {
		return nil
	}

	logger.Debugf("Validating config %s", kv)

	if kv.MspID != GlobalMSPID {
		return errors.Errorf("expecting MspID to be set to [%s] for DCAS config %s", GlobalMSPID, kv.Key)
	}

	if kv.AppVersion != DCASAppVersion {
		return errors.Errorf("unsupported DCAS config version %s for %s", kv.AppVersion, kv.Key)
	}

	var dcasCfg sidetreecfg.DCAS
	if err := unmarshal(kv.Value, &dcasCfg); err != nil {
		return errors.WithMessagef(err, "invalid config %s", kv.Key)
	}

	if dcasCfg.ChaincodeName == "" {
		return errors.Errorf("field 'ChaincodeName' is required for %s", kv.Key)
	}

	if dcasCfg.Collection == "" {
		return errors.Errorf("field 'Collection' is required for %s", kv.Key)
	}

	return nil
}
