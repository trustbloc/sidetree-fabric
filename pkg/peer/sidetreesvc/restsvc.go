/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package sidetreesvc

import (
	reqctx "context"

	"github.com/trustbloc/sidetree-core-go/pkg/api/protocol"
	"github.com/trustbloc/sidetree-core-go/pkg/dochandler"
	"github.com/trustbloc/sidetree-core-go/pkg/dochandler/didvalidator"
	"github.com/trustbloc/sidetree-core-go/pkg/processor"
	"github.com/trustbloc/sidetree-core-go/pkg/restapi/common"
	"github.com/trustbloc/sidetree-core-go/pkg/restapi/diddochandler"

	"github.com/trustbloc/sidetree-fabric/pkg/context/store"
	"github.com/trustbloc/sidetree-fabric/pkg/httpserver"
	"github.com/trustbloc/sidetree-fabric/pkg/peer/config"
	"github.com/trustbloc/sidetree-fabric/pkg/role"
)

type restService struct {
	listenURL  string
	httpServer *httpserver.Server
}

func newRESTService(cfg restConfig, handlers ...common.HTTPHandler) (*restService, error) {
	listenURL, err := cfg.SidetreeListenURL()
	if err != nil {
		return nil, err
	}

	var httpServer *httpserver.Server
	if len(handlers) > 0 {
		httpServer = httpserver.New(listenURL, handlers...)
	}

	return &restService{
		listenURL:  listenURL,
		httpServer: httpServer,
	}, nil
}

// Start starts the HTTP server if it is set
func (s *restService) Start() error {
	if s.httpServer != nil {
		logger.Infof("Starting REST service for Sidetree on [%s] ...", s.listenURL)

		return s.httpServer.Start()
	}

	return nil
}

// Stop stops the HTTP server if it is set
func (s *restService) Stop() {
	if s.httpServer != nil {
		logger.Infof("Stopping Sidetree REST service on [%s]", s.listenURL)

		if err := s.httpServer.Stop(reqctx.Background()); err != nil {
			logger.Warnf("Error stopping REST service: %s", err)
		}
	}
}

type restHandlers struct {
	channelID    string
	namespace    string
	httpHandlers []common.HTTPHandler
}

type protocolProvider interface {
	Protocol() protocol.Client
}

func newRESTHandlers(
	channelID string,
	cfg config.Namespace,
	dcasProvider dcasClientProvider,
	batchWriter dochandler.BatchWriter,
	protocolProvider protocolProvider) *restHandlers {

	if !role.IsResolver() && !role.IsBatchWriter() {
		return &restHandlers{
			channelID: channelID,
			namespace: cfg.Namespace,
		}
	}

	logger.Debugf("[%s] Creating document store for namespace [%s]", channelID, cfg.Namespace)

	opStore := store.New(channelID, cfg.Namespace, dcasProvider)

	// did document handler with did document validator for didDocNamespace
	didDocHandler := dochandler.New(
		cfg.Namespace,
		protocolProvider.Protocol(),
		didvalidator.New(opStore),
		batchWriter,
		processor.New(opStore),
	)

	var handlers []common.HTTPHandler

	if role.IsResolver() {
		logger.Debugf("Adding a Sidetree document resolver REST endpoint for namespace [%s].", cfg.Namespace)

		handlers = append(handlers, diddochandler.NewResolveHandler(cfg.BasePath, didDocHandler))
	}

	if role.IsBatchWriter() {
		logger.Debugf("Adding a Sidetree document update REST endpoint for namespace [%s].", cfg.Namespace)

		handlers = append(handlers, diddochandler.NewUpdateHandler(cfg.BasePath, didDocHandler))
	}

	return &restHandlers{
		channelID:    channelID,
		namespace:    cfg.Namespace,
		httpHandlers: handlers,
	}
}

// HTTPHandlers returns the HTTP handlers
func (h *restHandlers) HTTPHandlers() []common.HTTPHandler {
	return h.httpHandlers
}
