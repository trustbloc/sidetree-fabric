/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package sidetreesvc

import (
	"sync"

	gossipapi "github.com/hyperledger/fabric/extensions/gossip/api"
	"github.com/pkg/errors"
	ledgerconfig "github.com/trustbloc/fabric-peer-ext/pkg/config/ledgerconfig/config"
	"github.com/trustbloc/fabric-peer-ext/pkg/config/ledgerconfig/service"

	"github.com/trustbloc/sidetree-core-go/pkg/dochandler"
	"github.com/trustbloc/sidetree-core-go/pkg/restapi/common"

	"github.com/trustbloc/sidetree-fabric/pkg/config"
	ctxcommon "github.com/trustbloc/sidetree-fabric/pkg/context/common"
	"github.com/trustbloc/sidetree-fabric/pkg/context/store"
	"github.com/trustbloc/sidetree-fabric/pkg/observer/notifier"
	peerconfig "github.com/trustbloc/sidetree-fabric/pkg/peer/config"
	"github.com/trustbloc/sidetree-fabric/pkg/rest/authhandler"
	"github.com/trustbloc/sidetree-fabric/pkg/rest/blockchainhandler"
	"github.com/trustbloc/sidetree-fabric/pkg/rest/dcashandler"
	"github.com/trustbloc/sidetree-fabric/pkg/rest/filehandler"
	"github.com/trustbloc/sidetree-fabric/pkg/rest/sidetreehandler"
	"github.com/trustbloc/sidetree-fabric/pkg/role"
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
	notifier  *notifier.Notifier
	observer  *observerController
	contexts  map[string]*context
	handlers  map[string][]common.HTTPHandler
	cfgTxID   string
	txnChan   chan gossipapi.TxMetadata
}

func newChannelController(channelID string, providers *providers, configService config.SidetreeService, listener restServiceController) *channelController {
	ctrl := &channelController{
		providers:             providers,
		restServiceController: listener,
		channelID:             channelID,
		contexts:              make(map[string]*context),
		sidetreeCfgService:    configService,
		handlers:              make(map[string][]common.HTTPHandler),
	}

	if role.IsObserver() {
		ctrl.txnChan = make(chan gossipapi.TxMetadata, 1)
		ctrl.notifier = notifier.New(channelID, providers.BlockPublisher, ctrl.txnChan)
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

	for _, ctx := range c.contexts {
		ctx.Stop()
	}

	c.contexts = make(map[string]*context)

	if c.txnChan != nil {
		close(c.txnChan)
	}
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

	for _, h := range c.handlers {
		restHandlers = append(restHandlers, h...)
	}

	return restHandlers
}

func (c *channelController) load() error {
	logger.Debugf("[%s] Loading peer config for Sidetree", c.channelID)

	cfg, err := c.sidetreeCfgService.LoadSidetreePeer(c.PeerConfig.MSPID(), c.PeerConfig.PeerID())
	if err != nil {
		if errors.Cause(err) != service.ErrConfigNotFound {
			return err
		}

		logger.Info("No Sidetree configuration found for this peer.")
	}

	restHandlerCfg, err := c.loadRESTHandlerConfig()
	if err != nil {
		return err
	}

	dcasCfg, err := c.loadDCASConfig()
	if err != nil {
		return err
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	logger.Debugf("[%s] Updating Sidetree service channelController ...", c.channelID)

	storeProvider := store.NewProvider(c.channelID, c.sidetreeCfgService, c.DCASProvider)

	if err := c.loadContexts(restHandlerCfg.sidetree, dcasCfg, storeProvider); err != nil {
		return err
	}

	if err := c.loadRESTHandlers(restHandlerCfg); err != nil {
		return err
	}

	if err := c.restartObserver(cfg.Observer, dcasCfg, storeProvider); err != nil {
		return err
	}

	c.restServiceController.RestartRESTService()

	logger.Debugf("[%s] Successfully started Sidetree channelController.", c.channelID)

	return nil
}

type contextPair struct {
	newCtx *context
	oldCtx *context
}

func (c *channelController) loadContexts(handlers []sidetreehandler.Config, dcasCfg config.DCAS, storeProvider ctxcommon.OperationStoreProvider) error {
	loadedContexts, err := c.loadNewContexts(handlers, dcasCfg, storeProvider)
	if err != nil {
		return err
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
			return err
		}

		c.contexts[ctx.Namespace()] = ctx
	}

	return nil
}

func (c *channelController) loadNewContexts(handlers []sidetreehandler.Config, dcasCfg config.DCAS, storeProvider ctxcommon.OperationStoreProvider) ([]*context, error) {
	var contexts []*context

	for _, handlerCfg := range handlers {
		ctx, err := newContext(c.channelID, handlerCfg, dcasCfg, c.sidetreeCfgService, c.ContextProviders, storeProvider, c.RESTConfig)
		if err != nil {
			return nil, err
		}

		logger.Debugf("[%s] Loaded context for [%s]", c.channelID, handlerCfg.Namespace)

		contexts = append(contexts, ctx)
	}

	return contexts, nil
}

type restHandlerConfig struct {
	sidetree   []sidetreehandler.Config
	file       []filehandler.Config
	dcas       []dcashandler.Config
	blockchain []blockchainhandler.Config
}

func (c *channelController) loadRESTHandlerConfig() (*restHandlerConfig, error) {
	cfg := &restHandlerConfig{}

	var err error

	cfg.sidetree, err = c.loadSidetreeHandlerConfig()
	if err != nil {
		return nil, err
	}

	cfg.file, err = c.loadFileHandlerConfig()
	if err != nil {
		return nil, err
	}

	cfg.dcas, err = c.loadDCASHandlerConfig()
	if err != nil {
		return nil, err
	}

	cfg.blockchain, err = c.loadBlockchainHandlerConfig()
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *channelController) loadSidetreeHandlerConfig() ([]sidetreehandler.Config, error) {
	sidetreeHandlerCfg, err := c.sidetreeCfgService.LoadSidetreeHandlers(c.PeerConfig.MSPID(), c.PeerConfig.PeerID())
	if err == nil {
		return sidetreeHandlerCfg, nil
	}

	if errors.Cause(err) == service.ErrConfigNotFound {
		logger.Info("No Sidetree handler configuration found for this peer.")
		return nil, nil
	}

	return nil, err
}

func (c *channelController) loadFileHandlerConfig() ([]filehandler.Config, error) {
	fileHandlerCfg, err := c.sidetreeCfgService.LoadFileHandlers(c.PeerConfig.MSPID(), c.PeerConfig.PeerID())
	if err == nil {
		return fileHandlerCfg, nil
	}

	if errors.Cause(err) == service.ErrConfigNotFound {
		logger.Info("No file handler configuration found for this peer.")
		return nil, nil
	}

	return nil, err
}

func (c *channelController) loadDCASHandlerConfig() ([]dcashandler.Config, error) {
	dcasHandlerCfg, err := c.sidetreeCfgService.LoadDCASHandlers(c.PeerConfig.MSPID(), c.PeerConfig.PeerID())
	if err == nil {
		return dcasHandlerCfg, nil
	}

	if errors.Cause(err) == service.ErrConfigNotFound {
		logger.Info("No DCAS handler configuration found for this peer.")

		return nil, nil
	}

	return nil, err
}

func (c *channelController) loadBlockchainHandlerConfig() ([]blockchainhandler.Config, error) {
	blockchainHandlerCfg, err := c.sidetreeCfgService.LoadBlockchainHandlers(c.PeerConfig.MSPID(), c.PeerConfig.PeerID())
	if err == nil {
		return blockchainHandlerCfg, nil
	}

	if errors.Cause(err) == service.ErrConfigNotFound {
		logger.Info("No blockchain handler configuration found for this peer.")

		return nil, nil
	}

	return nil, err
}

func (c *channelController) loadDCASConfig() (config.DCAS, error) {
	return c.sidetreeCfgService.LoadDCAS()
}

func (c *channelController) restartObserver(observerCfg config.Observer, dcasCfg config.DCAS, storeProvider ctxcommon.OperationStoreProvider) error {
	if c.observer != nil {
		c.observer.Stop()
	}

	c.observer = newObserverController(c.channelID, c.PeerConfig, observerCfg, dcasCfg, c.ObserverProviders, storeProvider, c.txnChan)

	return c.observer.Start()
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
		logger.Debugf("Ignoring config update: %s", kv.Key)
		return
	}

	// If multiple components are updated in the same transaction then we'll get multiple notifications,
	// so avoid reloading the config multiple times by checking the ID of the last transaction that was handled.
	if !c.compareAndSetTxID(kv.TxID) {
		logger.Debugf("[%s] Got sidetree config update for %s but the update for TxID [%s] was already handled", c.channelID, kv.Key, kv.TxID)
		return
	}

	go func() {
		logger.Infof("[%s] Got config update for Sidetree: %s. Loading ...", c.channelID, kv.Key)

		if err := c.load(); err != nil {
			logger.Errorf("[%s] Error handling Sidetree config update: %s", c.channelID, err)
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
	if kv.MspID == c.PeerConfig.MSPID() && kv.PeerID == c.PeerConfig.PeerID() &&
		(kv.AppName == peerconfig.SidetreePeerAppName ||
			kv.AppName == peerconfig.FileHandlerAppName ||
			kv.AppName == peerconfig.DCASAppName ||
			kv.AppName == peerconfig.BlockchainHandlerAppName ||
			kv.AppName == peerconfig.DCASHandlerAppName) {
		return true
	}

	if kv.MspID == peerconfig.GlobalMSPID && c.isMonitoringNamespace(kv.AppName) {
		return true
	}

	return false
}

func (c *channelController) loadRESTHandlers(cfg *restHandlerConfig) error {
	if err := c.loadFileHandlers(cfg.file); err != nil {
		return err
	}

	c.loadDCASHandlers(cfg.dcas)

	c.loadBlockchainHandlers(cfg.blockchain)

	return nil
}

func (c *channelController) loadFileHandlers(handlerCfg []filehandler.Config) error {
	for _, cfg := range handlerCfg {
		h, err := c.loadFileHandler(cfg)
		if err != nil {
			return err
		}

		if h != nil {
			c.handlers[cfg.BasePath] = h.HTTPHandlers()
		}
	}

	return nil
}

func (c *channelController) loadFileHandler(cfg filehandler.Config) (*fileHandlers, error) {
	if !role.IsResolver() && !role.IsBatchWriter() {
		return nil, nil
	}

	docHandler, err := c.getDocHandler(cfg.IndexNamespace)
	if err != nil {
		return nil, errors.WithMessagef(err, "unable to get document handler for index document [%s]", cfg.IndexDocID)
	}

	handlers := &fileHandlers{}
	if role.IsResolver() && cfg.IndexDocID != "" {
		logger.Debugf("[%s] Adding file read handler for base path [%s]", c.channelID, cfg.BasePath)
		logger.Debugf("[%s] Authorization tokens for file read handler: %s", c.channelID, cfg.Authorization.ReadTokens)

		handlers.readHandler = c.authHandler(
			cfg.Authorization.ReadTokens,
			filehandler.NewRetrieveHandler(c.channelID, cfg, docHandler, c.DCASProvider),
		)
	}

	if role.IsBatchWriter() {
		logger.Debugf("[%s] Adding file upload handler for base path [%s]", c.channelID, cfg.BasePath)
		logger.Debugf("[%s] Authorization tokens for file upload handler: %s", c.channelID, cfg.Authorization.WriteTokens)

		handlers.writeHandler = c.authHandler(
			cfg.Authorization.WriteTokens,
			filehandler.NewUploadHandler(c.channelID, cfg, c.DCASProvider),
		)
	}

	return handlers, nil
}

func (c *channelController) loadDCASHandlers(handlerCfg []dcashandler.Config) {
	for _, cfg := range handlerCfg {
		c.handlers[cfg.BasePath] = c.loadDCASHandler(cfg)
	}
}

func (c *channelController) loadDCASHandler(cfg dcashandler.Config) []common.HTTPHandler {
	logger.Debugf("[%s] Adding DCAS handlers for base path [%s]", c.channelID, cfg.BasePath)
	logger.Debugf("[%s] Authorization tokens for DCAS handlers - read: %s, write: %s", c.channelID, cfg.Authorization.ReadTokens, cfg.Authorization.WriteTokens)

	return []common.HTTPHandler{
		c.authHandler(cfg.Authorization.ReadTokens, dcashandler.NewVersionHandler(c.channelID, cfg)),
		c.authHandler(cfg.Authorization.ReadTokens, dcashandler.NewRetrieveHandler(c.channelID, cfg, c.DCASProvider)),
		c.authHandler(cfg.Authorization.WriteTokens, dcashandler.NewUploadHandler(c.channelID, cfg, c.DCASProvider)),
	}
}

func (c *channelController) loadBlockchainHandlers(handlerCfg []blockchainhandler.Config) {
	for _, cfg := range handlerCfg {
		c.handlers[cfg.BasePath] = c.loadBlockchainHandler(cfg)
	}
}

func (c *channelController) loadBlockchainHandler(cfg blockchainhandler.Config) []common.HTTPHandler {
	logger.Debugf("[%s] Adding blockchain handlers for base path [%s]", c.channelID, cfg.BasePath)
	logger.Debugf("[%s] Authorization tokens for blockchain handlers: %s", c.channelID, cfg.Authorization.ReadTokens)

	readTokens := cfg.Authorization.ReadTokens

	return []common.HTTPHandler{
		c.authHandler(readTokens, blockchainhandler.NewTimeHandler(c.channelID, cfg, c.BlockchainProvider)),
		c.authHandler(readTokens, blockchainhandler.NewTimeByHashHandler(c.channelID, cfg, c.BlockchainProvider)),
		c.authHandler(readTokens, blockchainhandler.NewVersionHandler(c.channelID, cfg)),
		c.authHandler(readTokens, blockchainhandler.NewTransactionsSinceHandler(c.channelID, cfg, c.BlockchainProvider)),
		c.authHandler(readTokens, blockchainhandler.NewTransactionsHandler(c.channelID, cfg, c.BlockchainProvider)),
		c.authHandler(readTokens, blockchainhandler.NewFirstValidHandler(c.channelID, cfg, c.BlockchainProvider)),
		c.authHandler(readTokens, blockchainhandler.NewBlockByHashHandlerWithEncoding(c.channelID, cfg, c.BlockchainProvider)),
		c.authHandler(readTokens, blockchainhandler.NewBlockByHashHandler(c.channelID, cfg, c.BlockchainProvider)),
		c.authHandler(readTokens, blockchainhandler.NewBlocksFromNumHandlerWithEncoding(c.channelID, cfg, c.BlockchainProvider)),
		c.authHandler(readTokens, blockchainhandler.NewBlocksFromNumHandler(c.channelID, cfg, c.BlockchainProvider)),
		c.authHandler(readTokens, blockchainhandler.NewConfigBlockHandlerWithEncoding(c.channelID, cfg, c.BlockchainProvider)),
		c.authHandler(readTokens, blockchainhandler.NewConfigBlockHandler(c.channelID, cfg, c.BlockchainProvider)),
		c.authHandler(readTokens, blockchainhandler.NewConfigBlockByHashHandlerWithEncoding(c.channelID, cfg, c.BlockchainProvider)),
		c.authHandler(readTokens, blockchainhandler.NewConfigBlockByHashHandler(c.channelID, cfg, c.BlockchainProvider)),
	}
}

func (c *channelController) getDocHandler(ns string) (*dochandler.DocumentHandler, error) {
	ctx, ok := c.contexts[ns]
	if !ok {
		return nil, errors.Errorf("context not found for namespace [%s]", ns)
	}

	if ctx.rest == nil || ctx.rest.docHandler == nil {
		return nil, errors.Errorf("no document handler for namespace [%s]", ns)
	}

	return ctx.rest.docHandler, nil
}

// compareAndSetTxID sets the value of the transaction ID if it's not already set and returns true.
// If the transaction ID is already set then false is returned.
func (c *channelController) compareAndSetTxID(txID string) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.cfgTxID != txID {
		c.cfgTxID = txID
		return true
	}

	return false
}

func (c *channelController) authHandler(tokenNames []string, handler common.HTTPHandler) common.HTTPHandler {
	tokens := make([]string, len(tokenNames))

	for i, name := range tokenNames {
		tokens[i] = c.RESTConfig.SidetreeAPIToken(name)
	}

	return authhandler.New(c.channelID, tokens, handler)
}
