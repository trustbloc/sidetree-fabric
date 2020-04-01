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

// sidetreePeerValidator validates the SidetreePeer configuration
type sidetreePeerValidator struct {
}

func (v *sidetreePeerValidator) Validate(kv *config.KeyValue) error {
	if kv.AppName != SidetreePeerAppName {
		return nil
	}

	logger.Debugf("Validating config %s", kv)

	if kv.ComponentName != "" {
		return errors.Errorf("unexpected component [%s] for %s", kv.ComponentName, kv.Key)
	}

	if kv.PeerID == "" {
		return errors.Errorf("field PeerID required for %s", kv.Key)
	}

	if kv.AppVersion != SidetreePeerAppVersion {
		return errors.Errorf("unsupported application version [%s] for %s", kv.AppVersion, kv.Key)
	}

	var sidetreeCfg sidetreecfg.SidetreePeer
	if err := unmarshal(kv.Value, &sidetreeCfg); err != nil {
		return errors.WithMessagef(err, "invalid config %s", kv.Key)
	}

	if err := v.validateMonitor(kv, sidetreeCfg.Monitor); err != nil {
		return err
	}

	for _, ns := range sidetreeCfg.Namespaces {
		if err := v.validateNamespace(kv, ns); err != nil {
			return err
		}
	}

	return nil
}

func (v *sidetreePeerValidator) validateNamespace(kv *config.KeyValue, ns sidetreecfg.Namespace) error {
	if ns.Namespace == "" {
		return errors.Errorf("field 'Namespace' is required for %s", kv.Key)
	}

	if ns.BasePath == "" {
		return errors.Errorf("field 'BasePath' is required for %s", kv.Key)
	}

	if ns.BasePath[0:1] != "/" {
		return errors.Errorf("field 'BasePath' must begin with '/' for %s", kv.Key)
	}

	return nil
}

func (v *sidetreePeerValidator) validateMonitor(kv *config.KeyValue, cfg sidetreecfg.Monitor) error {
	if cfg.Period == 0 {
		logger.Infof("The Sidetree monitor period is set to 0 and therefore will be disabled for peer [%s].", kv.PeerID)
		return nil
	}

	if cfg.MetaDataChaincodeName == "" {
		return errors.Errorf("field 'MetaDataChaincodeName' is required for %s", kv.Key)
	}

	return nil
}
