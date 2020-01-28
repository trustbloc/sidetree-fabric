/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package sidetreesvc

import (
	"sync"

	ledgerconfig "github.com/trustbloc/fabric-peer-ext/pkg/config/ledgerconfig/config"
	"github.com/trustbloc/fabric-peer-ext/pkg/config/ledgerconfig/service"

	"github.com/trustbloc/sidetree-core-go/pkg/restapi/common"

	"github.com/trustbloc/sidetree-fabric/pkg/peer/config"
)

type restServiceController interface {
	RestartRESTService()
}

type channelController struct {
	*providers
	restServiceController
	sidetreeCfgService config.SidetreeService

	mutex     sync.RWMutex
	channelID string
	observer  *observerController
	monitor   *monitorController
	contexts  map[string]*context
}

func newChannelController(channelID string, providers *providers, configService config.SidetreeService, listener restServiceController) *channelController {
	ctrl := &channelController{
		providers:             providers,
		restServiceController: listener,
		channelID:             channelID,
		contexts:              make(map[string]*context),
		sidetreeCfgService:    configService,
	}

	providers.ConfigProvider.ForChannel(channelID).AddUpdateHandler(ctrl.handleUpdate)

	return ctrl
}

// Close frees all of the Sidetree resources for the channel
func (c *channelController) Close() {
	logger.Debugf("[%s] Closing Sidetree service channelController...", c.channelID)

	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.observer != nil {
		c.observer.Stop()
		c.observer = nil
	}

	if c.monitor != nil {
		c.monitor.Stop()
		c.monitor = nil
	}

	for _, ctx := range c.contexts {
		ctx.Stop()
	}

	c.contexts = make(map[string]*context)
}

// RESTHandlers returns the registered Sidetree REST handlers for the channel
func (c *channelController) RESTHandlers() []common.HTTPHandler {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	var restHandlers []common.HTTPHandler
	for _, ctx := range c.contexts {
		if ctx.rest != nil {
			restHandlers = append(restHandlers, ctx.rest.HTTPHandlers()...)
		}
	}

	return restHandlers
}

func (c *channelController) load() error {
	logger.Debugf("[%s] Loading peer sidetreeCfgService for Sidetree", c.channelID)

	cfg, err := c.sidetreeCfgService.LoadSidetreePeer(c.PeerConfig.MSPID(), c.PeerConfig.PeerID())
	if err != nil {
		if err == service.ErrConfigNotFound {
			// No Sidetree components defined for this peer. Stop all running channelController.
			logger.Info("No Sidetree configuration found for this peer.")
			c.Close()
			return nil
		}

		return err
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	logger.Debugf("[%s] Updating Sidetree service channelController ...", c.channelID)

	modified, err := c.loadContexts(cfg.Namespaces)
	if err != nil {
		return err
	}

	if c.observer == nil {
		c.observer = newObserverController(c.channelID, c.ObserverProviders)
		if err := c.observer.Start(); err != nil {
			return err
		}
	}

	if c.monitor == nil {
		c.monitor = newMonitorController(c.channelID, c.PeerConfig, cfg.Monitor, c.MonitorProviders)
		if err := c.monitor.Start(); err != nil {
			return err
		}
	}

	if modified {
		c.restServiceController.RestartRESTService()
	}

	logger.Debugf("[%s] Successfully started Sidetree channelController.", c.channelID)

	return nil
}

type contextPair struct {
	newCtx *context
	oldCtx *context
}

func (c *channelController) loadContexts(namespaces []config.Namespace) (modified bool, err error) {
	loadedContexts, err := c.loadNewContexts(namespaces)
	if err != nil {
		return false, err
	}

	var newContexts []*context
	var oldContexts []*context

	for _, ctxResult := range c.createContextMap(loadedContexts) {
		if ctxResult.oldCtx != nil {
			oldContexts = append(oldContexts, ctxResult.oldCtx)
		}

		if ctxResult.newCtx != nil {
			newContexts = append(newContexts, ctxResult.newCtx)
		}
	}

	for _, ctx := range oldContexts {
		ctx.Stop()

		delete(c.contexts, ctx.Namespace())
	}

	for _, ctx := range newContexts {
		if err := ctx.Start(); err != nil {
			return false, err
		}

		c.contexts[ctx.Namespace()] = ctx
	}

	return len(newContexts) > 0 || len(oldContexts) > 0, nil
}

func (c *channelController) loadNewContexts(namespaces []config.Namespace) ([]*context, error) {
	var contexts []*context

	for _, nsCfg := range namespaces {
		ctx, err := newContext(c.channelID, nsCfg, c.sidetreeCfgService, c.TxnProvider, c.DcasProvider)
		if err != nil {
			return nil, err
		}

		logger.Debugf("[%s] Loaded context for [%s]", c.channelID, nsCfg.Namespace)

		contexts = append(contexts, ctx)
	}

	return contexts, nil
}

func (c *channelController) createContextMap(newContexts []*context) map[string]*contextPair {
	contextMap := make(map[string]*contextPair)
	for _, ctx := range newContexts {
		contextMap[ctx.Namespace()] = &contextPair{newCtx: ctx}
	}

	for ns, existingCtx := range c.contexts {
		ctxPair, ok := contextMap[ns]
		if ok {
			ctxPair.oldCtx = existingCtx
		} else {
			contextMap[ns] = &contextPair{oldCtx: existingCtx}
		}
	}

	return contextMap
}

func (c *channelController) handleUpdate(kv *ledgerconfig.KeyValue) {
	if !c.shouldUpdate(kv) {
		logger.Debugf("Ignoring sidetreeCfgService update: %s", kv.Key)
		return
	}

	go func() {
		logger.Debugf("[%s] Got sidetreeCfgService update for Sidetree: %s. Loading ...", c.channelID, kv)

		if err := c.load(); err != nil {
			logger.Errorf("[%s] Error handling Sidetree sidetreeCfgService update: %s", c.channelID, err)
		} else {
			logger.Debugf("[%s] ... successfully updated Sidetree.", c.channelID)
		}
	}()
}

func (c *channelController) isMonitoringNamespace(namespace string) bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	for _, ctx := range c.contexts {
		if ctx.Namespace() == namespace {
			return true
		}
	}

	return false
}

func (c *channelController) shouldUpdate(kv *ledgerconfig.KeyValue) bool {
	if kv.MspID == c.PeerConfig.MSPID() && kv.PeerID == c.PeerConfig.PeerID() && kv.AppName == config.SidetreeAppName {
		return true
	}

	if kv.MspID == config.GlobalMSPID && c.isMonitoringNamespace(kv.AppName) {
		return true
	}

	return false
}
