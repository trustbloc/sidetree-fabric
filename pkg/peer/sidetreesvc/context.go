/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package sidetreesvc

import (
	"github.com/pkg/errors"

	"github.com/trustbloc/sidetree-core-go/pkg/dochandler"

	sidetreectx "github.com/trustbloc/sidetree-fabric/pkg/context"
	"github.com/trustbloc/sidetree-fabric/pkg/peer/config"
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

func newContext(channelID string, nsCfg config.Namespace, cfg config.SidetreeService, txnProvider txnServiceProvider, dcasProvider dcasClientProvider) (*context, error) {
	logger.Debugf("[%s] Creating Sidetree context for [%s]", channelID, nsCfg.Namespace)

	ctx, err := newSidetreeContext(channelID, nsCfg.Namespace, cfg, txnProvider, dcasProvider)
	if err != nil {
		return nil, err
	}

	logger.Debugf("[%s] Creating Sidetree batch writer for [%s]", channelID, nsCfg.Namespace)

	bw, err := newBatchWriter(channelID, nsCfg.Namespace, ctx, cfg)
	if err != nil {
		return nil, err
	}

	logger.Debugf("[%s] Creating Sidetree REST handlers [%s]", channelID, nsCfg.Namespace)

	restHandlers := newRESTHandlers(channelID, nsCfg, dcasProvider, bw, ctx)

	return &context{
		SidetreeContext: ctx,
		channelID:       channelID,
		batchWriter:     bw,
		rest:            restHandlers,
	}, nil
}

func newSidetreeContext(channelID, namespace string, cfg config.SidetreeService, txnProvider txnServiceProvider, dcasProvider dcasClientProvider) (*sidetreectx.SidetreeContext, error) {
	protocolVersions, err := cfg.LoadProtocols(namespace)
	if err != nil {
		return nil, err
	}

	if len(protocolVersions) == 0 {
		return nil, errors.Errorf("no protocols defined for [%s]", namespace)
	}

	return sidetreectx.New(channelID, namespace, protocolVersions, txnProvider, dcasProvider), nil
}
