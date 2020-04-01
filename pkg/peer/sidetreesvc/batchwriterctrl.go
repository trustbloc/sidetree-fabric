/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package sidetreesvc

import (
	"github.com/trustbloc/sidetree-core-go/pkg/batch"

	"github.com/trustbloc/sidetree-fabric/pkg/config"
	"github.com/trustbloc/sidetree-fabric/pkg/role"
)

type batchWriterController struct {
	*batch.Writer

	channelID string
	namespace string
}

func newBatchWriter(channelID, namespace string, ctx batch.Context, configService config.SidetreeService) (*batchWriterController, error) {
	var bw *batch.Writer
	if role.IsBatchWriter() {
		logger.Debugf("[%s] Creating Sidetree batch writer for [%s]", channelID, namespace)

		sidetreeCfg, err := configService.LoadSidetree(namespace)
		if err != nil {
			return nil, err
		}

		bw, err = batch.New(channelID+"_"+namespace, ctx, batch.WithBatchTimeout(sidetreeCfg.BatchWriterTimeout))
		if err != nil {
			return nil, err
		}
	}

	return &batchWriterController{
		Writer:    bw,
		channelID: channelID,
		namespace: namespace,
	}, nil
}

// Start starts the batch writer if it is set
func (bw *batchWriterController) Start() error {
	if bw.Writer != nil {
		logger.Infof("[%s] Starting batch writer for Sidetree [%s]", bw.channelID, bw.namespace)

		bw.Writer.Start()
	}
	return nil
}

// STop stops the batch writer if it is set
func (bw *batchWriterController) Stop() {
	if bw.Writer != nil {
		logger.Infof("[%s] Stopping batch writer for Sidetree [%s]", bw.channelID, bw.namespace)

		bw.Writer.Stop()
	}
}
