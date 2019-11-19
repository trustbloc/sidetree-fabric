/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package monitor

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/trustbloc/sidetree-core-go/pkg/api/batch"
	"github.com/trustbloc/sidetree-fabric/pkg/observer/common"
	"github.com/trustbloc/sidetree-fabric/pkg/observer/mocks"
)

func TestOperationStore_Put(t *testing.T) {
	dcasClient := mocks.NewMockDCASClient()
	dcasClientProvider := &mocks.DCASClientProvider{}
	dcasClientProvider.ForChannelReturns(dcasClient, nil)

	s := NewOperationStore(channel1, dcasClientProvider)
	require.NotNil(t, s)

	op1 := batch.Operation{
		ID: "op1",
	}
	op1Bytes, err := json.Marshal(op1)
	require.NoError(t, err)

	op2 := batch.Operation{
		ID: "op2",
	}

	_, err = dcasClient.Put(common.DocNs, common.DocColl, op1Bytes)
	require.NoError(t, err)
	require.NoError(t, s.Put([]batch.Operation{op1, op2}))

	op2Key, op2Bytes, err := common.MarshalDCAS(op2)
	require.NoError(t, err)
	opBytes, err := dcasClient.Get(common.DocNs, common.DocColl, op2Key)
	require.NoError(t, err)
	require.Equalf(t, op2Bytes, opBytes, "expecting that missing operation to be persisted in DCAS")
}

func TestOperationStore_PutError(t *testing.T) {
	dcasClient := mocks.NewMockDCASClient()
	dcasClientProvider := &mocks.DCASClientProvider{}
	dcasClientProvider.ForChannelReturns(dcasClient, nil)

	s := NewOperationStore(channel1, dcasClientProvider)
	op1 := batch.Operation{
		ID: "op1",
	}

	t.Run("DCAS get error", func(t *testing.T) {
		dcasClient.WithGetError(errors.New("injected DCAS error"))
		defer dcasClient.WithGetError(nil)
		require.Error(t, s.Put([]batch.Operation{op1}))
	})

	t.Run("DCAS put error", func(t *testing.T) {
		dcasClient.WithPutError(errors.New("injected DCAS error"))
		defer dcasClient.WithPutError(nil)
		require.Error(t, s.Put([]batch.Operation{op1}))
	})
}
