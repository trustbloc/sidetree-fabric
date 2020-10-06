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
	"github.com/trustbloc/sidetree-core-go/pkg/dochandler/transformer/didtransformer"
	"github.com/trustbloc/sidetree-core-go/pkg/processor"
	restcommon "github.com/trustbloc/sidetree-core-go/pkg/restapi/common"
	"github.com/trustbloc/sidetree-core-go/pkg/restapi/diddochandler"
	resthandler "github.com/trustbloc/sidetree-core-go/pkg/restapi/dochandler"

	"github.com/trustbloc/sidetree-fabric/pkg/common"
	"github.com/trustbloc/sidetree-fabric/pkg/config"
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

func newRESTService(cfg restConfig, handlers ...restcommon.HTTPHandler) (*restService, error) {
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

type tokenProvider interface {
	SidetreeAPIToken(name string) string
}

func newRESTHandlers(
	channelID string,
	cfg sidetreehandler.Config,
	batchWriter dochandler.BatchWriter,
	pc protocol.Client,
	opStore processor.OperationStoreClient,
	tokenProvider tokenProvider,
	configService config.SidetreeService) (*restHandlers, error) {

	if !role.IsResolver() && !role.IsBatchWriter() {
		return &restHandlers{
			channelID: channelID,
			namespace: cfg.Namespace,
		}, nil
	}

	logger.Debugf("[%s] Creating document store for namespace [%s]", channelID, cfg.Namespace)

	getTransformer, getResolveHandler, getUpdateHandler, err := newProviders(cfg.DocType)
	if err != nil {
		return nil, err
	}

	sidetreeCfg, err := configService.LoadSidetree(cfg.Namespace)
	if err != nil {
		return nil, err
	}

	docHandler := dochandler.New(
		cfg.Namespace,
		cfg.Aliases,
		pc,
		getTransformer(sidetreeCfg),
		batchWriter,
		processor.New(channelID+"_"+cfg.Namespace, opStore, pc),
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
			newEndpoint("/operations", authhandler.New(channelID, authTokens(cfg.Authorization.WriteTokens, tokenProvider), getUpdateHandler(cfg, docHandler, pc))),
		)
	}

	service.endpoints = append(service.endpoints,
		newEndpoint("/version", sidetreehandler.NewVersionHandler(channelID, cfg, pc)),
	)

	return &restHandlers{
		channelID:  channelID,
		namespace:  cfg.Namespace,
		service:    service,
		docHandler: docHandler,
	}, nil
}

type validatorProvider func(cfg config.Sidetree) dochandler.DocumentTransformer
type resolveHandlerProvider func(sidetreehandler.Config, resthandler.Resolver) restcommon.HTTPHandler
type updateHandlerProvider func(sidetreehandler.Config, resthandler.Processor, protocol.Client) restcommon.HTTPHandler

var (
	didDocTransformerProvider = func(cfg config.Sidetree) dochandler.DocumentTransformer {
		return didtransformer.New(didtransformer.WithMethodContext(cfg.MethodContext))
	}

	didDocResolveProvider = func(cfg sidetreehandler.Config, resolver resthandler.Resolver) restcommon.HTTPHandler {
		return diddochandler.NewResolveHandler(cfg.BasePath, resolver)
	}

	didDocUpdateProvider = func(cfg sidetreehandler.Config, processor resthandler.Processor, pc protocol.Client) restcommon.HTTPHandler {
		return diddochandler.NewUpdateHandler(cfg.BasePath, processor, pc)
	}

	fileTransformerProvider = func(cfg config.Sidetree) dochandler.DocumentTransformer {
		return filehandler.NewTransformer()
	}

	fileResolveProvider = func(cfg sidetreehandler.Config, resolver resthandler.Resolver) restcommon.HTTPHandler {
		return filehandler.NewResolveIndexHandler(cfg.BasePath, resolver)
	}

	fileUpdateProvider = func(cfg sidetreehandler.Config, processor resthandler.Processor, pc protocol.Client) restcommon.HTTPHandler {
		return filehandler.NewUpdateIndexHandler(cfg.BasePath, processor, pc)
	}
)

func newProviders(docType common.DocumentType) (validatorProvider, resolveHandlerProvider, updateHandlerProvider, error) {
	switch docType {
	case common.FileIndexType:
		return fileTransformerProvider, fileResolveProvider, fileUpdateProvider, nil
	case common.DIDDocType:
		return didDocTransformerProvider, didDocResolveProvider, didDocUpdateProvider, nil
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
