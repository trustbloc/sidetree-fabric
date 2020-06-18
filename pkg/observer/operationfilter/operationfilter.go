/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package operationfilter

import (
	sidetreeobserver "github.com/trustbloc/sidetree-core-go/pkg/observer"
	"github.com/trustbloc/sidetree-core-go/pkg/processor"
	"github.com/trustbloc/sidetree-fabric/pkg/context/common"
)

// Provider implements an operation filter provider
type Provider struct {
	channelID       string
	opStoreProvider common.OperationStoreProvider
	pcp             common.ProtocolClientProvider
}

// NewProvider returns a new operation filter provider
func NewProvider(channelID string, opStoreProvider common.OperationStoreProvider, pcp common.ProtocolClientProvider) *Provider {
	return &Provider{
		channelID:       channelID,
		opStoreProvider: opStoreProvider,
		pcp:             pcp,
	}
}

// Get returns the operation filter for the given namespace
func (f *Provider) Get(namespace string) (sidetreeobserver.OperationFilter, error) {
	opStore, err := f.opStoreProvider.ForNamespace(namespace)
	if err != nil {
		return nil, err
	}

	pc, err := f.pcp.ForNamespace(namespace)
	if err != nil {
		return nil, err
	}

	return processor.NewOperationFilter(f.channelID+"_"+namespace, opStore, pc), nil
}
