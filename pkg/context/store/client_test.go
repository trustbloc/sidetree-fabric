/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package store

import (
	"fmt"
	"testing"

	"github.com/hyperledger/fabric-protos-go/ledger/queryresult"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"github.com/trustbloc/sidetree-core-go/pkg/api/batch"
	"github.com/trustbloc/sidetree-core-go/pkg/docutil"
	stmocks "github.com/trustbloc/sidetree-fabric/pkg/mocks"
	obmocks "github.com/trustbloc/sidetree-fabric/pkg/observer/mocks"
)

const (
	chID = "mychannel"
	id   = "id"

	namespace = "did:sidetree"
)

func TestNew(t *testing.T) {
	dcasProvider := &stmocks.DCASClientProvider{}
	dcasClient := obmocks.NewMockDCASClient()
	dcasProvider.ForChannelReturns(dcasClient, nil)

	c := New(chID, namespace, dcasProvider)
	require.NotNil(t, c)
}

func TestProviderError(t *testing.T) {
	testErr := errors.New("provider error")
	dcasProvider := &stmocks.DCASClientProvider{}
	dcasProvider.ForChannelReturns(nil, testErr)

	c := New(chID, namespace, dcasProvider)
	require.NotNil(t, c)

	payload, err := c.Get(id)
	require.EqualError(t, err, testErr.Error())
	require.Nil(t, payload)
}

func TestWriteContent(t *testing.T) {
	dcasProvider := &stmocks.DCASClientProvider{}
	dcasClient := obmocks.NewMockDCASClient()
	dcasProvider.ForChannelReturns(dcasClient, nil)

	didID := namespace + docutil.NamespaceDelimiter + id

	vk1 := &queryresult.KV{
		Namespace: documentCC + "~" + collection,
		Key:       didID,
		Value:     []byte("{}"),
	}

	query := fmt.Sprintf(queryByIDTemplate, didID)
	dcasClient.WithQueryResults(documentCC, collection, query, []*queryresult.KV{vk1})
	c := New(chID, namespace, dcasProvider)

	ops, err := c.Get(id)
	require.Nil(t, err)
	require.NotNil(t, ops)
	require.Equal(t, 1, len(ops))
}

func TestGetOperationsError(t *testing.T) {

	doc, err := getOperations([][]byte{[]byte("[test : 123]")})
	require.NotNil(t, err)
	require.Nil(t, doc)
	require.Contains(t, err.Error(), "invalid character")
}

func TestClient_Put(t *testing.T) {
	c := New(chID, namespace, &stmocks.DCASClientProvider{})

	require.PanicsWithValue(t, "not implemented", func() {
		err := c.Put(&batch.Operation{})
		require.NoError(t, err)
	})
}

func TestSort(t *testing.T) {
	var operations []*batch.Operation

	const testID = "id"
	delete := &batch.Operation{ID: testID, Type: "delete", TransactionTime: 2, TransactionNumber: 1}
	update := &batch.Operation{ID: testID, Type: "update", TransactionTime: 1, TransactionNumber: 7}
	create := &batch.Operation{ID: testID, Type: "create", TransactionTime: 1, TransactionNumber: 1}

	operations = append(operations, delete)
	operations = append(operations, update)
	operations = append(operations, create)

	result := sortChronologically(operations)
	require.Equal(t, create.Type, result[0].Type)
	require.Equal(t, update.Type, result[1].Type)
	require.Equal(t, delete.Type, result[2].Type)
}
