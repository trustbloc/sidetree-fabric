/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package sidetreesvc

import (
	reqctx "context"

	"github.com/pkg/errors"

	"github.com/trustbloc/sidetree-core-go/pkg/api/protocol"
	"github.com/trustbloc/sidetree-core-go/pkg/dochandler"
	"github.com/trustbloc/sidetree-core-go/pkg/dochandler/didvalidator"
	"github.com/trustbloc/sidetree-core-go/pkg/dochandler/docvalidator"
	"github.com/trustbloc/sidetree-core-go/pkg/processor"
	"github.com/trustbloc/sidetree-core-go/pkg/restapi/common"
	"github.com/trustbloc/sidetree-core-go/pkg/restapi/diddochandler"
	resthandler "github.com/trustbloc/sidetree-core-go/pkg/restapi/dochandler"

	"github.com/trustbloc/sidetree-fabric/pkg/filehandler"
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
	docHandler   *dochandler.DocumentHandler
}

type protocolProvider interface {
	Protocol() protocol.Client
}

func newRESTHandlers(
	channelID string,
	cfg config.Namespace,
	batchWriter dochandler.BatchWriter,
	protocolProvider protocolProvider,
	opStore processor.OperationStoreClient) (*restHandlers, error) {

	if !role.IsResolver() && !role.IsBatchWriter() {
		return &restHandlers{
			channelID: channelID,
			namespace: cfg.Namespace,
		}, nil
	}

	logger.Debugf("[%s] Creating document store for namespace [%s]", channelID, cfg.Namespace)

	getValidator, getResolveHandler, getUpdateHandler, err := newProviders(cfg.DocType)
	if err != nil {
		return nil, err
	}

	docHandler := dochandler.New(
		cfg.Namespace,
		protocolProvider.Protocol(),
		getValidator(opStore),
		batchWriter,
		processor.New(channelID+"_"+cfg.Namespace, opStore),
	)

	var handlers []common.HTTPHandler

	if role.IsResolver() {
		logger.Debugf("Adding a Sidetree document resolver REST endpoint for namespace [%s].", cfg.Namespace)

		handlers = append(handlers, getResolveHandler(cfg, docHandler))
	}

	if role.IsBatchWriter() {
		logger.Debugf("Adding a Sidetree document update REST endpoint for namespace [%s].", cfg.Namespace)

		handlers = append(handlers, getUpdateHandler(cfg, docHandler))
	}

	return &restHandlers{
		channelID:    channelID,
		namespace:    cfg.Namespace,
		httpHandlers: handlers,
		docHandler:   docHandler,
	}, nil
}

// HTTPHandlers returns the HTTP handlers
func (h *restHandlers) HTTPHandlers() []common.HTTPHandler {
	return h.httpHandlers
}

type validatorProvider func(opStore docvalidator.OperationStoreClient) dochandler.DocumentValidator
type resolveHandlerProvider func(config.Namespace, resthandler.Resolver) common.HTTPHandler
type updateHandlerProvider func(config.Namespace, resthandler.Processor) common.HTTPHandler

var (
	didDocValidatorProvider = func(opStore docvalidator.OperationStoreClient) dochandler.DocumentValidator {
		return didvalidator.New(opStore)
	}

	didDocResolveProvider = func(cfg config.Namespace, resolver resthandler.Resolver) common.HTTPHandler {
		return diddochandler.NewResolveHandler(cfg.BasePath, resolver)
	}

	didDocUpdateProvider = func(cfg config.Namespace, processor resthandler.Processor) common.HTTPHandler {
		return diddochandler.NewUpdateHandler(cfg.BasePath, processor)
	}

	fileValidatorProvider = func(opStore docvalidator.OperationStoreClient) dochandler.DocumentValidator {
		return filehandler.NewValidator(opStore)
	}

	fileResolveProvider = func(cfg config.Namespace, resolver resthandler.Resolver) common.HTTPHandler {
		return newFileIdxResolveHandler(cfg.BasePath, resolver)
	}

	fileUpdateProvider = func(cfg config.Namespace, processor resthandler.Processor) common.HTTPHandler {
		return newFileIdxUpdateHandler(cfg.BasePath, processor)
	}
)

func newProviders(docType config.DocumentType) (validatorProvider, resolveHandlerProvider, updateHandlerProvider, error) {
	switch docType {
	case config.FileIndexType:
		return fileValidatorProvider, fileResolveProvider, fileUpdateProvider, nil
	case config.DIDDocType:
		return didDocValidatorProvider, didDocResolveProvider, didDocUpdateProvider, nil
	default:
		return nil, nil, nil, errors.Errorf("unsupported document type [%s]", docType)
	}
}
