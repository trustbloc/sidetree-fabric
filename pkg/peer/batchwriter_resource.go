/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package peer

import (
	"time"

	"github.com/trustbloc/sidetree-core-go/pkg/batch"
	"github.com/trustbloc/sidetree-fabric/pkg/role"
)

type batchWriterResource struct {
	*batch.Writer
}

type batchWriterConfig interface {
	GetBatchTimeout() time.Duration
}

func newBatchWriter(context batch.Context, config batchWriterConfig) *batchWriterResource {
	bw := &batchWriterResource{}

	if role.IsBatchWriter() {
		logger.Infof("Starting batch writer")

		var err error
		bw.Writer, err = batch.New(context, batch.WithBatchTimeout(config.GetBatchTimeout()))
		if err != nil {
			panic(err)
		}

		bw.Start()
	}

	return bw
}

// Close stops the batch writer
func (r *batchWriterResource) Close() {
	if r.Writer != nil {
		logger.Infof("Stopping batch writer")
		r.Stop()
	}
}
