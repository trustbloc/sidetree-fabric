/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package sidetreesvc

import (
	"sync"

	"github.com/hyperledger/fabric/common/flogging"
	dcas "github.com/trustbloc/fabric-peer-ext/pkg/collections/offledger/dcas/client"
	ledgerconfig "github.com/trustbloc/fabric-peer-ext/pkg/config/ledgerconfig/config"
	txnapi "github.com/trustbloc/fabric-peer-ext/pkg/txn/api"
	"github.com/trustbloc/sidetree-core-go/pkg/batch/cutter"
	"github.com/trustbloc/sidetree-core-go/pkg/restapi/common"

	"github.com/trustbloc/sidetree-fabric/pkg/config"
	ctxcommon "github.com/trustbloc/sidetree-fabric/pkg/context/common"
	"github.com/trustbloc/sidetree-fabric/pkg/observer"
	"github.com/trustbloc/sidetree-fabric/pkg/peer/discovery"
)

var logger = flogging.MustGetLogger("sidetree_peer")

type txnServiceProvider interface {
	ForChannel(channelID string) (txnapi.Service, error)
}

type configServiceProvider interface {
	ForChannel(channelID string) ledgerconfig.Service
}

type peerConfig interface {
	PeerID() string
	PeerAddress() string
	MSPID() string
}

type dcasClientProvider interface {
	ForChannel(channelID string) (dcas.DCAS, error)
}

type restConfig interface {
	SidetreeListenURL() (string, error)
	SidetreeListenPort() int
	SidetreeTLSCertificate() string
	SidetreeTLSKey() string
	SidetreeAPIToken(name string) string
}

type sidetreeConfigProvider interface {
	ForChannel(channelID string) config.SidetreeService
}

type operationQueueProvider interface {
	Create(channelID string, namespace string) (cutter.OperationQueue, error)
}

type discoveryProvider interface {
	UpdateLocalServicesForChannel(channelID string, services []discovery.Service)
	ServicesForChannel(channelID string) []discovery.Service
}

type providers struct {
	*ContextProviders

	PeerConfig        peerConfig
	RESTConfig        restConfig
	ConfigProvider    configServiceProvider
	ObserverProviders *observer.ClientProviders
	BlockPublisher    ctxcommon.BlockPublisherProvider
	DiscoveryProvider discoveryProvider
}

// Provider implements a Sidetree services provider which is responsible for managing Sidetree
// services for all channels that a peer has joined.
type Provider struct {
	*providers

	sidetreeConfigProvider sidetreeConfigProvider
	restService            *restService
	chanControllers        map[string]*channelController
	mutex                  sync.RWMutex
}

// NewProvider returns a new Sidetree services provider
func NewProvider(providers *providers, sidetreeConfigProvider sidetreeConfigProvider) *Provider {
	logger.Info("Creating Sidetree services provider")

	return &Provider{
		providers:              providers,
		sidetreeConfigProvider: sidetreeConfigProvider,
		chanControllers:        make(map[string]*channelController),
	}
}

// ChannelJoined is invoked when a peer joins a channel. A new channel manager is created to manage the
// Sidetree resources for the channel.
func (p *Provider) ChannelJoined(channelID string) {
	logger.Infof("[%s] Joined channel.", channelID)

	sidetreeCfgService := p.sidetreeConfigProvider.ForChannel(channelID)

	ctrl := newChannelController(channelID, p.providers, sidetreeCfgService, p)

	go func() {
		if err := ctrl.load(); err != nil {
			logger.Warnf("[%s] Error loading Sidetree service channelController: %s. The configuration may not yet be loaded. Once the configuration is loaded, the Sidetree service channelController will be started.", channelID, err)
		}
	}()

	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.chanControllers[channelID] = ctrl
}

// Close closes all Sidetree resources
func (p *Provider) Close() {
	logger.Infof("Closing Sidetree resources ...")

	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.restService != nil {
		p.restService.Stop()
		p.restService = nil
	}

	for _, service := range p.chanControllers {
		service.Close()
	}

	logger.Infof("... successfully closed Sidetree resources.")
}

// RestartRESTService restarts the REST server
func (p *Provider) RestartRESTService() {
	go func() {
		p.restartRESTService()
	}()
}

func (p *Provider) restartRESTService() {
	logger.Info("Attempting to restart REST service ...")

	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.restService != nil {
		p.restService.Stop()
		p.restService = nil
	}

	var handlers []common.HTTPHandler
	for _, mgr := range p.chanControllers {
		handlersForChannel := mgr.RESTHandlers()

		logger.Debugf("... adding %d REST handlers for channel [%s]", len(handlersForChannel), mgr.channelID)

		handlers = append(handlers, handlersForChannel...)
	}

	if len(handlers) == 0 {
		logger.Info("... not starting REST service since no handlers have been defined.")
		return
	}

	restService, err := newRESTService(p.RESTConfig, handlers...)
	if err != nil {
		logger.Errorf("Unable to create REST service: %s", err)
	}

	p.restService = restService

	if err := p.restService.Start(); err != nil {
		logger.Errorf("Unable to start REST service: %s", err)
	}
}
