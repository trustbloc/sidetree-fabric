/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package config

import (
	"github.com/pkg/errors"

	"github.com/trustbloc/fabric-peer-ext/pkg/config/ledgerconfig/config"
	"github.com/trustbloc/sidetree-core-go/pkg/docutil"

	sidetreecfg "github.com/trustbloc/sidetree-fabric/pkg/config"
	"github.com/trustbloc/sidetree-fabric/pkg/observer"
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

	// Validation of protocol depends on the version.
	switch kv.ComponentVersion {
	default:
		// For now, only one protocol version exists
		return v.validateProtocolV0(kv)
	}
}

func (v *sidetreeValidator) validateProtocolV0(kv *config.KeyValue) error {
	p, err := unmarshalProtocol(kv.Value)
	if err != nil {
		return errors.WithMessagef(err, "invalid protocol config for %s", kv.Key)
	}

	if _, err := docutil.GetHash(p.HashAlgorithmInMultiHashCode); err != nil {
		return errors.WithMessagef(err, "error in Sidetree protocol for %s", kv.Key)
	}

	if p.MaxOperationsPerBatch == 0 {
		return errors.Errorf("field 'MaxOperationsPerBatch' must contain a value greater than 0 for %s", kv.Key)
	}

	if p.MaxDeltaByteSize == 0 {
		return errors.Errorf("field 'MaxDeltaByteSize' must contain a value greater than 0 for %s", kv.Key)
	}

	return nil
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
