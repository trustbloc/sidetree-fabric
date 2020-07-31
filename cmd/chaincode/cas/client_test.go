/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
package cas

import (
	"crypto"
	"encoding/base64"
	"errors"
	"testing"

	"github.com/trustbloc/fabric-peer-ext/pkg/collections/offledger/dcas"
	"github.com/trustbloc/sidetree-core-go/pkg/docutil"
	"github.com/trustbloc/sidetree-fabric/cmd/chaincode/mocks"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/stretchr/testify/require"
)

const collection = "docs"

func TestWrite(t *testing.T) {

	client := getClient()

	t.Run("Success", func(t *testing.T) {
		content := getOperationBytes(getCreateOperation())
		addr, err := client.Write(content)
		require.Nil(t, err)
		require.NotNil(t, addr)
		require.Equal(t, encodedSHA256Hash(content), addr)

		payload, err := client.Read(addr)
		require.Nil(t, err)
		require.NotNil(t, payload)
		require.Equal(t, content, payload)

		// test write same content
		addr2, err := client.Write(content)
		require.Nil(t, err)
		require.Equal(t, addr, addr2)
	})

	t.Run("Marshal error", func(t *testing.T) {
		reset := dcas.SetJSONMarshaller(func(m map[string]interface{}) (bytes []byte, e error) {
			return nil, errors.New("injected marshal error")
		})
		defer reset()

		content := getOperationBytes(getCreateOperation())
		addr, err := client.Write(content)
		require.Error(t, err)
		require.Empty(t, addr)
	})
}

func TestWrite_PutPrivateError(t *testing.T) {

	testErr := errors.New("write error")
	mockStub := newMockStub()
	mockStub.PutPrivateErr = testErr
	client := New(mockStub, collection)

	content := getOperationBytes(getCreateOperation())
	address, err := client.Write(content)
	require.NotNil(t, err)
	require.Empty(t, address)
	require.Contains(t, err.Error(), testErr.Error())
}

func TestRead(t *testing.T) {

	client := getClient()

	content := getOperationBytes(getCreateOperation())
	addr, err := client.Write(content)
	require.Nil(t, err)
	require.NotNil(t, addr)

	read, err := client.Read(addr)
	require.Nil(t, err)
	require.NotNil(t, read)
	require.Equal(t, read, content)

	read, err = client.Read("non-existent")
	require.Nil(t, err)
	require.Nil(t, read)
}

func TestQuery(t *testing.T) {
	const query = "{\"selector\":{\"unique_suffix\":\"1234\"},\"use_index\":[\"_design/indexUniqueSuffixDoc\",\"indexUniqueSuffix\"]}"

	client := getClient()

	t.Run("Success", func(t *testing.T) {
		content := getOperationBytes(getCreateOperation())
		addr, err := client.Write(content)
		require.Nil(t, err)
		require.NotNil(t, addr)

		read, err := client.Query(query)
		require.Nil(t, err)
		require.NotNil(t, read)
	})

	t.Run("Query error", func(t *testing.T) {
		client.stub.(*mocks.MockStub).GetPrivateQueryErr = errors.New("injected query error")
		results, err := client.Query(query)
		require.Error(t, err)
		require.Nil(t, results)
	})
}

func TestRead_GetPrivateError(t *testing.T) {

	testErr := errors.New("read error")
	mockStub := newMockStub()
	mockStub.GetPrivateErr = testErr
	client := New(mockStub, collection)

	payload, err := client.Read("address")
	require.NotNil(t, err)
	require.Nil(t, payload)
	require.Contains(t, err.Error(), testErr.Error())
}

func encodedSHA256Hash(bytes []byte) string {
	h := crypto.SHA256.New()
	if _, err := h.Write(bytes); err != nil {
		panic(err)
	}

	return base64.URLEncoding.EncodeToString(h.Sum(nil))
}

func newMockStub() *mocks.MockStub {
	return mocks.NewMockStub("mockcc", &mockCC{})
}

func newClient() *Client {
	return New(newMockStub(), collection)
}

func newClientWithTx() *Client {
	client := newClient()
	client.stub.(*mocks.MockStub).MockTransactionStart("txID")
	return client
}

type mockCC struct {
}

func (cc *mockCC) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

func (cc *mockCC) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

func getClient() *Client {
	return newClientWithTx()
}

func getOperationBytes(op *Operation) []byte {

	bytes, err := docutil.MarshalCanonical(op)
	if err != nil {
		panic(err)
	}

	return bytes
}

func getCreateOperation() *Operation {
	return &Operation{UniqueSuffix: "abc", Type: "create"}
}

// Operation defines sample operation
type Operation struct {
	Type         string `json:"type"`
	UniqueSuffix string `json:"unique_suffix"`
}
