/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package monitor

import (
	"github.com/pkg/errors"
	"github.com/trustbloc/fabric-peer-ext/pkg/collections/offledger/dcas/client"
	"github.com/trustbloc/sidetree-core-go/pkg/api/batch"
	"github.com/trustbloc/sidetree-fabric/pkg/observer/common"
)

// OperationStore ensures that a given set of operations is persisted in the Document DCAS store
type OperationStore struct {
	dcasClientProvider common.DCASClientProvider
	channelID          string
}

// NewOperationStore returns an OperationStore
func NewOperationStore(channelID string, dcasClientProvider common.DCASClientProvider) *OperationStore {
	return &OperationStore{
		channelID:          channelID,
		dcasClientProvider: dcasClientProvider,
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
	key, opBytes, err := common.MarshalDCAS(op)
	if err != nil {
		return newMonitorError(errors.Wrapf(err, "failed to get DCAS key and value for operation [%s]", op.ID), false)
	}

	dcasClient, err := s.dcasClient()
	if err != nil {
		return newMonitorError(err, true)
	}

	retrievedBytes, err := dcasClient.Get(common.DocNs, common.DocColl, key)
	if err != nil {
		return newMonitorError(errors.Wrapf(err, "failed to retrieve operation [%s] by key [%s]", op.ID, key), true)
	}

	if len(retrievedBytes) == 0 {
		logger.Infof("[%s] Operation [%s] was not found in DCAS using key [%s]. Persisting...", s.channelID, op.ID, key)
		if _, err = dcasClient.Put(common.DocNs, common.DocColl, opBytes); err != nil {
			return newMonitorError(errors.Wrapf(err, "failed to persist operation [%s]", op.ID), true)
		}
	} else {
		logger.Debugf("[%s] Operation [%s] was found in DCAS using key [%s]", s.channelID, op.ID, key)
	}

	return nil
}

func (s *OperationStore) dcasClient() (client.DCAS, error) {
	return s.dcasClientProvider.ForChannel(s.channelID)
}
