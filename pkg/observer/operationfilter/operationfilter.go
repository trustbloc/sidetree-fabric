/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package operationfilter

import (
	sidetreeobserver "github.com/trustbloc/sidetree-core-go/pkg/observer"
	"github.com/trustbloc/sidetree-core-go/pkg/processor"

	"github.com/trustbloc/sidetree-fabric/pkg/observer/common"
)

// Provider implements an operation filter provider
type Provider struct {
	channelID             string
	opStoreClientProvider common.OperationStoreClientProvider
}

// NewProvider returns a new operation filter provider
func NewProvider(channelID string, opStoreClientProvider common.OperationStoreClientProvider) *Provider {
	return &Provider{
		channelID:             channelID,
		opStoreClientProvider: opStoreClientProvider,
	}
}

// Get returns the operation filter for the given namespace
func (f *Provider) Get(namespace string) sidetreeobserver.OperationFilter {
	opStoreClient := f.opStoreClientProvider.Get(f.channelID, namespace)

	return processor.NewOperationFilter(f.channelID+"_"+namespace, opStoreClient)
}
