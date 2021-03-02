/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package config

import (
	"github.com/pkg/errors"
	"github.com/trustbloc/fabric-peer-ext/pkg/config/ledgerconfig/config"
	"github.com/trustbloc/sidetree-fabric/pkg/protocolversion/versions/v1_0/validator"

	sidetreecfg "github.com/trustbloc/sidetree-fabric/pkg/config"
	"github.com/trustbloc/sidetree-fabric/pkg/observer"
	"github.com/trustbloc/sidetree-fabric/pkg/protocolversion"
	"github.com/trustbloc/sidetree-fabric/pkg/protocolversion/common"
)

const (
	sidetreeTag = "sidetree"
)

// sidetreeValidator validates the Sidetree configuration including Protocols
type sidetreeValidator struct {
}

func (v *sidetreeValidator) Validate(kv *config.KeyValue) error {
	if !canValidate(kv) {
		return nil
	}

	logger.Debugf("Validating config %s", kv)

	if kv.MspID != GlobalMSPID {
		return errors.Errorf("expecting MspID to be set to [%s] for Sidetree config %s", GlobalMSPID, kv.Key)
	}

	switch kv.ComponentName {
	case "":
		return v.validateConfig(kv)
	case ProtocolComponentName:
		return v.validateProtocol(kv)
	default:
		return errors.Errorf("unexpected component [%s] for %s", kv.ComponentName, kv.Key)
	}
}

func (v *sidetreeValidator) validateConfig(kv *config.KeyValue) error {
	logger.Debugf("Validating Sidetree config %s", kv)

	if kv.AppVersion != SidetreeAppVersion {
		return errors.Errorf("unsupported application version [%s] for %s", kv.AppVersion, kv.Key)
	}

	var sidetreeCfg sidetreecfg.Sidetree
	if err := unmarshal(kv.Value, &sidetreeCfg); err != nil {
		return errors.WithMessagef(err, "invalid config %s", kv.Key)
	}

	if sidetreeCfg.BatchWriterTimeout == 0 {
		return errors.Errorf("field 'BatchWriterTimeout' must contain a value greater than 0 for %s", kv.Key)
	}

	if sidetreeCfg.ChaincodeName == "" {
		return errors.Errorf("field 'ChaincodeName' is required for %s", kv.Key)
	}

	if sidetreeCfg.Collection == "" {
		return errors.Errorf("field 'Collection' is required for %s", kv.Key)
	}

	if sidetreeCfg.Collection == observer.MetaDataColName {
		return errors.Errorf("field 'Collection' must not use reserved name [%s] for %s", sidetreeCfg.Collection, kv.Key)
	}

	return nil
}

func (v *sidetreeValidator) validateProtocol(kv *config.KeyValue) error {
	logger.Debugf("Validating Sidetree Protocol config %s", kv)

	p, err := unmarshalProtocol(kv.Value)
	if err != nil {
		return errors.WithMessagef(err, "invalid protocol config for %s", kv.Key)
	}

	// Validation of protocol depends on the version.
	if common.Version(kv.ComponentVersion).Matches(protocolversion.V0_1) {
		err = validator.Validate(p)
		if err != nil {
			return errors.WithMessagef(err, "invalid protocol config for %s", kv.Key)
		}

		return nil
	}

	return errors.Errorf("protocol version [%s] not supported for %s", kv.ComponentVersion, kv.Key)
}

func canValidate(kv *config.KeyValue) bool {
	for _, tag := range kv.Tags {
		if tag == sidetreeTag {
			logger.Debugf("Found tag [%s] in %s", tag, kv)
			return true
		}
	}

	logger.Debugf("Did not find tag [%s] in %s", sidetreeTag, kv)

	return false
}
