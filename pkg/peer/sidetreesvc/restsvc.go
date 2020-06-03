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

	"github.com/trustbloc/sidetree-fabric/pkg/httpserver"
	"github.com/trustbloc/sidetree-fabric/pkg/rest/authhandler"
	"github.com/trustbloc/sidetree-fabric/pkg/rest/filehandler"
	"github.com/trustbloc/sidetree-fabric/pkg/rest/sidetreehandler"
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
		httpServer = httpserver.New(listenURL, cfg.SidetreeTLSCertificate(), cfg.SidetreeTLSKey(), handlers...)
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
	channelID  string
	namespace  string
	service    *service
	docHandler *dochandler.DocumentHandler
}

type protocolProvider interface {
	Protocol() protocol.Client
}

type tokenProvider interface {
	SidetreeAPIToken(name string) string
}

func newRESTHandlers(
	channelID string,
	cfg sidetreehandler.Config,
	batchWriter dochandler.BatchWriter,
	protocolProvider protocolProvider,
	opStore processor.OperationStoreClient,
	tokenProvider tokenProvider) (*restHandlers, error) {

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

	service := newService(cfg.Namespace, apiVersion, cfg.BasePath)

	if role.IsResolver() {
		logger.Debugf("[%s] Adding a Sidetree document resolver REST endpoint for namespace [%s].", channelID, cfg.Namespace)
		logger.Debugf("[%s] Authorization tokens for document resolver REST endpoint for namespace [%s]: %s", channelID, cfg.Namespace, cfg.Authorization.ReadTokens)

		service.endpoints = append(service.endpoints,
			newEndpoint("/identifiers", authhandler.New(channelID, authTokens(cfg.Authorization.ReadTokens, tokenProvider), getResolveHandler(cfg, docHandler))),
		)
	}

	if role.IsBatchWriter() {
		logger.Debugf("[%s] Adding a Sidetree document update REST endpoint for namespace [%s].", channelID, cfg.Namespace)
		logger.Debugf("[%s] Authorization tokens for document update REST endpoint for namespace [%s]: %s", channelID, cfg.Namespace, cfg.Authorization.WriteTokens)

		service.endpoints = append(service.endpoints,
			newEndpoint("/operations", authhandler.New(channelID, authTokens(cfg.Authorization.WriteTokens, tokenProvider), getUpdateHandler(cfg, docHandler))),
		)
	}

	service.endpoints = append(service.endpoints,
		newEndpoint("/version", sidetreehandler.NewVersionHandler(channelID, cfg)),
	)

	return &restHandlers{
		channelID:  channelID,
		namespace:  cfg.Namespace,
		service:    service,
		docHandler: docHandler,
	}, nil
}

type validatorProvider func(opStore docvalidator.OperationStoreClient) dochandler.DocumentValidator
type resolveHandlerProvider func(sidetreehandler.Config, resthandler.Resolver) common.HTTPHandler
type updateHandlerProvider func(sidetreehandler.Config, resthandler.Processor) common.HTTPHandler

var (
	didDocValidatorProvider = func(opStore docvalidator.OperationStoreClient) dochandler.DocumentValidator {
		return didvalidator.New(opStore)
	}

	didDocResolveProvider = func(cfg sidetreehandler.Config, resolver resthandler.Resolver) common.HTTPHandler {
		return diddochandler.NewResolveHandler(cfg.BasePath, resolver)
	}

	didDocUpdateProvider = func(cfg sidetreehandler.Config, processor resthandler.Processor) common.HTTPHandler {
		return diddochandler.NewUpdateHandler(cfg.BasePath, processor)
	}

	fileValidatorProvider = func(opStore docvalidator.OperationStoreClient) dochandler.DocumentValidator {
		return filehandler.NewValidator(opStore)
	}

	fileResolveProvider = func(cfg sidetreehandler.Config, resolver resthandler.Resolver) common.HTTPHandler {
		return filehandler.NewResolveIndexHandler(cfg.BasePath, resolver)
	}

	fileUpdateProvider = func(cfg sidetreehandler.Config, processor resthandler.Processor) common.HTTPHandler {
		return filehandler.NewUpdateIndexHandler(cfg.BasePath, processor)
	}
)

func newProviders(docType sidetreehandler.DocumentType) (validatorProvider, resolveHandlerProvider, updateHandlerProvider, error) {
	switch docType {
	case sidetreehandler.FileIndexType:
		return fileValidatorProvider, fileResolveProvider, fileUpdateProvider, nil
	case sidetreehandler.DIDDocType:
		return didDocValidatorProvider, didDocResolveProvider, didDocUpdateProvider, nil
	default:
		return nil, nil, nil, errors.Errorf("unsupported document type [%s]", docType)
	}
}

func authTokens(names []string, tokenProvider tokenProvider) []string {
	tokens := make([]string, len(names))

	for i, name := range names {
		tokens[i] = tokenProvider.SidetreeAPIToken(name)
	}

	return tokens
}
