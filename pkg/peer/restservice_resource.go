/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package peer

import (
	"context"

	"github.com/trustbloc/sidetree-core-go/pkg/api/protocol"
	"github.com/trustbloc/sidetree-core-go/pkg/dochandler"
	"github.com/trustbloc/sidetree-core-go/pkg/dochandler/didvalidator"
	"github.com/trustbloc/sidetree-core-go/pkg/processor"
	"github.com/trustbloc/sidetree-core-go/pkg/restapi/common"
	"github.com/trustbloc/sidetree-core-go/pkg/restapi/diddochandler"
	"github.com/trustbloc/sidetree-fabric/pkg/httpserver"
	"github.com/trustbloc/sidetree-fabric/pkg/role"
)

type restServiceConfig interface {
	GetListenURL() string
}

type protocolProvider interface {
	Protocol() protocol.Client
}

type restService struct {
	*httpserver.Server
}

func newRESTService(
	didDocStore processor.OperationStoreClient,
	batchWriter dochandler.BatchWriter,
	config restServiceConfig,
	protocolProvider protocolProvider) *restService {

	logger.Info("Starting REST service..")

	// did document handler with did document validator for didDocNamespace
	didDocHandler := dochandler.New(
		didDocNamespace,
		protocolProvider.Protocol(),
		didvalidator.New(didDocStore),
		batchWriter,
		processor.New(didDocStore),
	)

	var handlers []common.HTTPHandler
	if role.IsResolver() {
		logger.Info("Adding a Sidetree document resolver REST endpoint.")
		handlers = append(handlers, diddochandler.NewResolveHandler(didDocHandler))
	}
	if role.IsBatchWriter() {
		logger.Info("Adding a Sidetree document update REST endpoint.")
		handlers = append(handlers, diddochandler.NewUpdateHandler(didDocHandler))
	}

	restSvc := httpserver.New(config.GetListenURL(), handlers...)

	err := restSvc.Start()
	if err != nil {
		panic(err)
	}

	return &restService{
		Server: restSvc,
	}
}

// Close stops the REST server
func (r *restService) Close() {
	logger.Infof("Stopping REST service")

	if err := r.Stop(context.Background()); err != nil {
		logger.Warnf("Error stopping REST server: %s", err)
	}
}
