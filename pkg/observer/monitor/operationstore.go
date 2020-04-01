/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package monitor

import (
	"strings"

	"github.com/pkg/errors"

	"github.com/trustbloc/sidetree-core-go/pkg/api/batch"
	"github.com/trustbloc/sidetree-core-go/pkg/observer"

	ctxcommon "github.com/trustbloc/sidetree-fabric/pkg/context/common"
	"github.com/trustbloc/sidetree-fabric/pkg/observer/common"
)

// OperationStoreProvider manages operation store providers by namespace
type OperationStoreProvider struct {
	channelID       string
	opStoreProvider ctxcommon.OperationStoreProvider
}

// NewOperationStoreProvider returns a new operation store provider
func NewOperationStoreProvider(channelID string, opStoreProvider ctxcommon.OperationStoreProvider) *OperationStoreProvider {
	return &OperationStoreProvider{
		channelID:       channelID,
		opStoreProvider: opStoreProvider,
	}
}

// ForNamespace returns the operation store for the given namespace
func (p *OperationStoreProvider) ForNamespace(namespace string) (observer.OperationStore, error) {
	opStore, err := p.opStoreProvider.ForNamespace(namespace)
	if err != nil {
		return nil, newMonitorError(errors.Wrapf(err, "unable to get operation store for namespace [%s]", namespace), true)
	}

	return NewOperationStore(p.channelID, opStore), nil
}

// OperationStore ensures that a given set of operations is persisted in the Document DCAS store
type OperationStore struct {
	opStore   ctxcommon.OperationStore
	channelID string
}

// NewOperationStore returns an OperationStore
func NewOperationStore(channelID string, opStore ctxcommon.OperationStore) *OperationStore {
	return &OperationStore{
		channelID: channelID,
		opStore:   opStore,
	}
}

// Put first checks if the given operations have already been persisted; if not, then they will be persisted.
func (s *OperationStore) Put(ops []*batch.Operation) error {
	for _, op := range ops {
		if err := s.checkOperation(op); err != nil {
			return err
		}
	}
	return nil
}

func (s *OperationStore) checkOperation(op *batch.Operation) error {
	key, _, err := common.MarshalDCAS(op)
	if err != nil {
		return newMonitorError(errors.Wrapf(err, "failed to get DCAS key and value for operation [%s]", op.ID), false)
	}

	retrievedOps, err := s.opStore.Get(key)
	if err != nil {
		// TODO: Should have a well-defined 'not found' error
		if !strings.Contains(err.Error(), "not found") {
			return newMonitorError(errors.Wrapf(err, "failed to retrieve operation [%s] by key [%s]", op.ID, key), true)
		}
	}

	if len(retrievedOps) == 0 {
		logger.Infof("[%s] Operation [%s] was not found in DCAS using key [%s]. Persisting...", s.channelID, op.ID, key)
		if err = s.opStore.Put([]*batch.Operation{op}); err != nil {
			return newMonitorError(errors.Wrapf(err, "failed to persist operation [%s]", op.ID), true)
		}
	} else {
		logger.Debugf("[%s] Operation [%s] was found in DCAS using key [%s]", s.channelID, op.ID, key)
	}

	return nil
}
