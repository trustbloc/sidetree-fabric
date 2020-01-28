/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package config

import (
	"bytes"

	"github.com/hyperledger/fabric/common/flogging"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	ledgerconfig "github.com/trustbloc/fabric-peer-ext/pkg/config/ledgerconfig/config"
	protocolApi "github.com/trustbloc/sidetree-core-go/pkg/api/protocol"
)

var logger = flogging.MustGetLogger("sidetree_peer")

type configServiceProvider interface {
	ForChannel(channelID string) ledgerconfig.Service
}

// SidetreeProvider manages Sidetree configuration for the various channels
type SidetreeProvider struct {
	configProvider configServiceProvider
}

// SidetreeService is a service that loads Sidetree configuration
type SidetreeService interface {
	LoadProtocols(namespace string) (map[string]protocolApi.Protocol, error)
	LoadSidetree(namespace string) (Sidetree, error)
	LoadSidetreePeer(mspID, peerID string) (SidetreePeer, error)
}

// NewSidetreeProvider returns a new SidetreeProvider instance
func NewSidetreeProvider(configProvider configServiceProvider) *SidetreeProvider {
	logger.Info("Creating Sidetree config provider")
	return &SidetreeProvider{configProvider: configProvider}
}

// ForChannel returns the service for the given channel
func (p *SidetreeProvider) ForChannel(channelID string) SidetreeService {
	return &sidetreeService{service: p.configProvider.ForChannel(channelID)}
}

type sidetreeService struct {
	service ledgerconfig.Service
}

// LoadSidetree loads the Sidetree configuration for the given namespace
func (c *sidetreeService) LoadSidetree(namespace string) (Sidetree, error) {
	key := ledgerconfig.NewAppKey(GlobalMSPID, namespace, "1")

	var sidetreeConfig Sidetree
	if err := c.load(key, &sidetreeConfig); err != nil {
		return Sidetree{}, errors.WithMessagef(err, "unable to load Sidetree config key %s", key)
	}

	if sidetreeConfig.BatchWriterTimeout == 0 {
		return Sidetree{}, errors.New("batchWriterTimeout must be greater than 0")
	}

	return sidetreeConfig, nil
}

// LoadSidetreePeer loads the peer-specific Sidetree configuration
func (c *sidetreeService) LoadSidetreePeer(mspID, peerID string) (SidetreePeer, error) {
	key := ledgerconfig.NewPeerKey(mspID, peerID, SidetreeAppName, SidetreeAppVersion)

	var sidetreeConfig SidetreePeer
	if err := c.load(key, &sidetreeConfig); err != nil {
		return SidetreePeer{}, errors.WithMessagef(err, "unable to load Sidetree peer config key %s", key)
	}

	return sidetreeConfig, nil
}

// LoadProtocols loads the Sidetree protocols for the given namespace
func (c *sidetreeService) LoadProtocols(namespace string) (map[string]protocolApi.Protocol, error) {
	criteria := &ledgerconfig.Criteria{
		MspID:         GlobalMSPID,
		AppName:       namespace,
		AppVersion:    "1",
		ComponentName: ProtocolComponentName,
	}

	results, err := c.service.Query(criteria)
	if err != nil {
		return nil, errors.WithMessagef(err, "error loading Sidetree protocol config for criteria %s", criteria)
	}

	protocolVersions := make(map[string]protocolApi.Protocol)

	for _, protoCfg := range results {
		protocol, err := unmarshalProtocol(protoCfg.Value)
		if err != nil {
			return nil, errors.WithMessagef(err, "error unmarshalling Sidetree protocol config for instance [%s]", namespace)
		}

		protocolVersions[protoCfg.ComponentVersion] = *protocol
	}

	return protocolVersions, nil
}

func (c *sidetreeService) load(key *ledgerconfig.Key, v interface{}) error {
	cfg, err := c.service.Get(key)
	if err != nil {
		return errors.WithMessagef(err, "error getting Sidetree config for key %s", key)
	}

	vp := viper.New()
	vp.SetConfigType(string(cfg.Format))

	if err := vp.ReadConfig(bytes.NewBufferString(cfg.Config)); err != nil {
		return errors.WithMessage(err, "error reading config")
	}

	if err := vp.Unmarshal(v); err != nil {
		return errors.WithMessage(err, "error unmarshalling config")
	}

	return nil
}

func unmarshalProtocol(cfg *ledgerconfig.Value) (*protocolApi.Protocol, error) {
	v := viper.New()
	v.SetConfigType(string(cfg.Format))

	err := v.ReadConfig(bytes.NewBufferString(cfg.Config))
	if err != nil {
		return nil, errors.WithMessage(err, "error reading Sidetree protocol config")
	}

	protocol := &protocolApi.Protocol{}
	if err := v.Unmarshal(protocol); err != nil {
		return nil, errors.WithMessage(err, "error unmarshalling Sidetree protocol config")
	}

	return protocol, nil
}
