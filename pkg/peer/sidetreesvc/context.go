/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package sidetreesvc

import (
	"github.com/pkg/errors"
	casApi "github.com/trustbloc/sidetree-core-go/pkg/api/cas"
	protocolApi "github.com/trustbloc/sidetree-core-go/pkg/api/protocol"
	"github.com/trustbloc/sidetree-core-go/pkg/dochandler"

	bcclient "github.com/trustbloc/sidetree-fabric/pkg/client"
	"github.com/trustbloc/sidetree-fabric/pkg/common"
	"github.com/trustbloc/sidetree-fabric/pkg/config"
	sidetreectx "github.com/trustbloc/sidetree-fabric/pkg/context"
	"github.com/trustbloc/sidetree-fabric/pkg/context/cas"
	ctxcommon "github.com/trustbloc/sidetree-fabric/pkg/context/common"
	"github.com/trustbloc/sidetree-fabric/pkg/rest/sidetreehandler"
)

type batchWriter interface {
	dochandler.BatchWriter

	Start() error
	Stop()
}

type context struct {
	*sidetreectx.SidetreeContext

	channelID   string
	batchWriter batchWriter
	rest        *restHandlers
}

// BatchWriter returns the batch writer
func (c *context) BatchWriter() dochandler.BatchWriter {
	return c.batchWriter
}

// Start starts the Sidetree resources held by the context
func (c *context) Start() error {
	logger.Debugf("[%s] Starting Sidetree [%s]", c.channelID, c.Namespace())

	return c.batchWriter.Start()
}

// Stop stops the Sidetree resources held by the context
func (c *context) Stop() {
	logger.Debugf("[%s] Stopping Sidetree [%s]", c.channelID, c.Namespace())

	c.batchWriter.Stop()
}

type blockchainClientProvider interface {
	ForChannel(channelID string) (bcclient.Blockchain, error)
}

type protocolVersionFactory interface {
	CreateProtocolVersion(version string, p protocolApi.Protocol, casClient casApi.Client, opStore ctxcommon.OperationStore, docType common.DocumentType) (protocolApi.Version, error)
}

// ContextProviders defines the providers required by the context
type ContextProviders struct {
	*sidetreectx.Providers
	BlockchainProvider blockchainClientProvider
	VersionFactory     protocolVersionFactory
}

func newContext(channelID string, handlerCfg sidetreehandler.Config, dcasCfg config.DCAS, cfg config.SidetreeService, providers *ContextProviders, opStoreProvider ctxcommon.OperationStoreProvider, tokenProvider tokenProvider) (*context, error) {
	logger.Debugf("[%s] Creating Sidetree context for [%s]", channelID, handlerCfg.Namespace)

	dcasClient, err := providers.DCASProvider.ForChannel(channelID)
	if err != nil {
		return nil, err
	}

	ctx, err := newSidetreeContext(channelID, handlerCfg.Namespace, cfg, handlerCfg.DocType, dcasCfg, opStoreProvider, cas.New(dcasCfg, dcasClient), providers)
	if err != nil {
		return nil, err
	}

	logger.Debugf("[%s] Creating Sidetree batch writer for [%s]", channelID, handlerCfg.Namespace)

	bw, err := newBatchWriter(channelID, handlerCfg.Namespace, ctx, cfg)
	if err != nil {
		return nil, err
	}

	logger.Debugf("[%s] Creating Sidetree REST handlers [%s]", channelID, handlerCfg.Namespace)

	store, err := opStoreProvider.ForNamespace(handlerCfg.Namespace)
	if err != nil {
		return nil, err
	}

	restHandlers, err := newRESTHandlers(channelID, handlerCfg, bw, ctx.Protocol(), store, tokenProvider, cfg)
	if err != nil {
		return nil, err
	}

	return &context{
		SidetreeContext: ctx,
		channelID:       channelID,
		batchWriter:     bw,
		rest:            restHandlers,
	}, nil
}

func newSidetreeContext(channelID, namespace string, cfg config.SidetreeService, docType common.DocumentType, dcasCfg config.DCAS, opStoreProvider ctxcommon.OperationStoreProvider, casClient casApi.Client, providers *ContextProviders) (*sidetreectx.SidetreeContext, error) {
	protocols, err := cfg.LoadProtocols(namespace)
	if err != nil {
		return nil, err
	}

	if len(protocols) == 0 {
		return nil, errors.Errorf("no protocols defined for [%s]", namespace)
	}

	opStore, err := opStoreProvider.ForNamespace(namespace)
	if err != nil {
		return nil, err
	}

	var protocolVersions []protocolApi.Version
	for version, p := range protocols {
		pv, err := providers.VersionFactory.CreateProtocolVersion(version, p, casClient, opStore, docType)
		if err != nil {
			// This may be a case where support for a protocol version has been removed but the protocol is still in the ledger config.
			// Log an error but continue processing other protocol versions.
			logger.Errorf("[%s] Error creating protocol version [%s] for namespace [%s]: %s", channelID, namespace, version, err)

			continue
		}

		protocolVersions = append(protocolVersions, pv)
	}

	return sidetreectx.New(channelID, namespace, dcasCfg, casClient, protocolVersions, providers.Providers)
}
