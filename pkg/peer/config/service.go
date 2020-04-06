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

	"github.com/trustbloc/sidetree-fabric/pkg/config"
	"github.com/trustbloc/sidetree-fabric/pkg/filehandler"
)

var logger = flogging.MustGetLogger("sidetree_peer")

type configServiceProvider interface {
	ForChannel(channelID string) ledgerconfig.Service
}

type validatorRegistry interface {
	Register(v ledgerconfig.Validator)
}

// SidetreeProvider manages Sidetree configuration for the various channels
type SidetreeProvider struct {
	configProvider configServiceProvider
}

// NewSidetreeProvider returns a new SidetreeProvider instance
func NewSidetreeProvider(configProvider configServiceProvider, registry validatorRegistry) *SidetreeProvider {
	logger.Info("Creating Sidetree config provider")

	registry.Register(&sidetreeValidator{})
	registry.Register(&sidetreePeerValidator{})
	registry.Register(&fileHandlerValidator{})
	registry.Register(&dcasValidator{})

	return &SidetreeProvider{
		configProvider: configProvider,
	}
}

// ForChannel returns the service for the given channel
func (p *SidetreeProvider) ForChannel(channelID string) config.SidetreeService {
	return &sidetreeService{service: p.configProvider.ForChannel(channelID)}
}

type sidetreeService struct {
	service ledgerconfig.Service
}

// LoadSidetree loads the Sidetree configuration for the given namespace
func (c *sidetreeService) LoadSidetree(namespace string) (config.Sidetree, error) {
	key := ledgerconfig.NewAppKey(GlobalMSPID, namespace, SidetreeAppVersion)

	var sidetreeConfig config.Sidetree
	if err := c.load(key, &sidetreeConfig); err != nil {
		return config.Sidetree{}, errors.WithMessagef(err, "unable to load Sidetree config key %s", key)
	}

	if sidetreeConfig.BatchWriterTimeout == 0 {
		return config.Sidetree{}, errors.New("batchWriterTimeout must be greater than 0")
	}

	return sidetreeConfig, nil
}

// LoadSidetreePeer loads the peer-specific Sidetree configuration
func (c *sidetreeService) LoadSidetreePeer(mspID, peerID string) (config.SidetreePeer, error) {
	key := ledgerconfig.NewPeerKey(mspID, peerID, SidetreePeerAppName, SidetreePeerAppVersion)

	var sidetreeConfig config.SidetreePeer
	if err := c.load(key, &sidetreeConfig); err != nil {
		return config.SidetreePeer{}, errors.WithMessagef(err, "unable to load Sidetree peer config key %s", key)
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

// LoadFileHandlers loads the file handler configuration
func (c *sidetreeService) LoadFileHandlers(mspID, peerID string) ([]filehandler.Config, error) {
	criteria := &ledgerconfig.Criteria{
		MspID:      mspID,
		PeerID:     peerID,
		AppName:    FileHandlerAppName,
		AppVersion: FileHandlerAppVersion,
	}

	results, err := c.service.Query(criteria)
	if err != nil {
		return nil, errors.WithMessagef(err, "error loading file handler config for criteria %s", criteria)
	}

	var handlers []filehandler.Config
	for _, kv := range results {
		h := filehandler.Config{}
		if err := unmarshal(kv.Value, &h); err != nil {
			return nil, err
		}

		handlers = append(handlers, h)
	}

	return handlers, nil
}

// LoadDCAS loads the DCAS configuration
func (c *sidetreeService) LoadDCAS() (config.DCAS, error) {
	key := ledgerconfig.NewAppKey(GlobalMSPID, DCASAppName, SidetreeAppVersion)

	var dcasCfg config.DCAS
	if err := c.load(key, &dcasCfg); err != nil {
		return config.DCAS{}, errors.WithMessagef(err, "unable to load DCAS config key %s", key)
	}

	if dcasCfg.ChaincodeName == "" {
		return config.DCAS{}, errors.New("field 'ChaincodeName' is required")
	}

	if dcasCfg.Collection == "" {
		return config.DCAS{}, errors.New("field 'Collection' is required")
	}

	return dcasCfg, nil
}

func (c *sidetreeService) load(key *ledgerconfig.Key, v interface{}) error {
	cfg, err := c.service.Get(key)
	if err != nil {
		return err
	}

	return unmarshal(cfg, v)
}

func unmarshal(value *ledgerconfig.Value, v interface{}) error {
	vp := viper.New()
	vp.SetConfigType(string(value.Format))

	if err := vp.ReadConfig(bytes.NewBufferString(value.Config)); err != nil {
		return errors.WithMessage(err, "error reading config")
	}

	if err := vp.Unmarshal(v); err != nil {
		return errors.WithMessage(err, "error unmarshalling config")
	}

	return nil
}

func unmarshalProtocol(cfg *ledgerconfig.Value) (*protocolApi.Protocol, error) {
	protocol := &protocolApi.Protocol{}

	if err := unmarshal(cfg, protocol); err != nil {
		return nil, errors.WithMessage(err, "error unmarshalling Sidetree protocol config")
	}

	return protocol, nil
}
