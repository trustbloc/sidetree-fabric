/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package sidetreesvc

import (
	"github.com/pkg/errors"

	"github.com/trustbloc/sidetree-core-go/pkg/dochandler"

	bcclient "github.com/trustbloc/sidetree-fabric/pkg/client"
	"github.com/trustbloc/sidetree-fabric/pkg/config"
	sidetreectx "github.com/trustbloc/sidetree-fabric/pkg/context"
	"github.com/trustbloc/sidetree-fabric/pkg/context/common"
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

// Start stops the Sidetree resources held by the context
func (c *context) Stop() {
	logger.Debugf("[%s] Stopping Sidetree [%s]", c.channelID, c.Namespace())

	c.batchWriter.Stop()
}

type blockchainClientProvider interface {
	ForChannel(channelID string) (bcclient.Blockchain, error)
}

// ContextProviders defines the providers required by the context
type ContextProviders struct {
	TxnProvider            txnServiceProvider
	DCASProvider           dcasClientProvider
	OperationQueueProvider operationQueueProvider
	BlockchainProvider     blockchainClientProvider
}

func newContext(channelID string, nsCfg config.Namespace, dcasCfg config.DCAS, cfg config.SidetreeService, providers *ContextProviders, opStoreProvider common.OperationStoreProvider) (*context, error) {
	logger.Debugf("[%s] Creating Sidetree context for [%s]", channelID, nsCfg.Namespace)

	ctx, err := newSidetreeContext(channelID, nsCfg.Namespace, cfg, dcasCfg, providers.TxnProvider, providers.DCASProvider, providers.OperationQueueProvider)
	if err != nil {
		return nil, err
	}

	logger.Debugf("[%s] Creating Sidetree batch writer for [%s]", channelID, nsCfg.Namespace)

	bw, err := newBatchWriter(channelID, nsCfg.Namespace, ctx, cfg)
	if err != nil {
		return nil, err
	}

	logger.Debugf("[%s] Creating Sidetree REST handlers [%s]", channelID, nsCfg.Namespace)

	store, err := opStoreProvider.ForNamespace(nsCfg.Namespace)
	if err != nil {
		return nil, err
	}

	restHandlers, err := newRESTHandlers(channelID, nsCfg, bw, ctx, store)
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

func newSidetreeContext(channelID, namespace string, cfg config.SidetreeService, dcasCfg config.DCAS, txnProvider txnServiceProvider, dcasProvider dcasClientProvider, opQueueProvider operationQueueProvider) (*sidetreectx.SidetreeContext, error) {
	protocolVersions, err := cfg.LoadProtocols(namespace)
	if err != nil {
		return nil, err
	}

	if len(protocolVersions) == 0 {
		return nil, errors.Errorf("no protocols defined for [%s]", namespace)
	}

	return sidetreectx.New(channelID, namespace, dcasCfg, protocolVersions, txnProvider, dcasProvider, opQueueProvider)
}
