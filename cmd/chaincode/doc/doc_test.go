/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package doc

import (
	"crypto"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/trustbloc/sidetree-core-go/pkg/docutil"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/trustbloc/sidetree-fabric/cmd/chaincode/mocks"
)

const (
	testID = "did:sidetree:abc"
	ccName = "document_cc"
)

func TestNew(t *testing.T) {
	req := require.New(t)

	cc := New(ccName)
	req.NotNil(cc)

	req.Equal(ccName, cc.Name())
	req.Equal(ccVersion, cc.Version())
	req.Equal(cc, cc.Chaincode())

	dbArtifacts := cc.GetDBArtifacts()
	req.NotNil(dbArtifacts)

	artifact, ok := dbArtifacts[couchDB]
	req.True(ok)
	req.Empty(artifact.Indexes)
	req.Len(artifact.CollectionIndexes, 1)
}

func TestInvoke(t *testing.T) {

	stub := prepareStub()

	testInvalidFunctionName(t, stub)

	// run last
	checkInit(t, stub, [][]byte{})
}

func TestWrite(t *testing.T) {

	stub := prepareStub()

	testPayload := getOperationBytes(getCreateOperation())
	address, err := invoke(stub, [][]byte{[]byte(write), testPayload})
	assert.Nil(t, err)
	assert.Equal(t, encodedSHA256Hash(testPayload), string(address))

	payload, err := invoke(stub, [][]byte{[]byte(read), []byte(address)})
	assert.Nil(t, err)
	assert.Equal(t, testPayload, payload)
}

func TestWriteError(t *testing.T) {

	testErr := fmt.Errorf("write error")
	stub := prepareStub()
	stub.PutPrivateErr = testErr

	testPayload := getOperationBytes(getCreateOperation())

	payload, err := invoke(stub, [][]byte{[]byte(write), testPayload})
	assert.NotNil(t, err)
	assert.Nil(t, payload)
	assert.Contains(t, err.Error(), testErr.Error())
}

func TestWrite_MissingContent(t *testing.T) {

	stub := prepareStub()

	address, err := invoke(stub, [][]byte{[]byte(write), []byte("")})
	assert.NotNil(t, err)
	assert.Nil(t, address)
	assert.Contains(t, err.Error(), "missing content")
}

func TestRead(t *testing.T) {

	stub := prepareStub()

	testPayload := getOperationBytes(getCreateOperation())
	testPayloadAddress := encodedSHA256Hash(testPayload)

	payload, err := invoke(stub, [][]byte{[]byte(read), []byte(testPayloadAddress)})
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "not found")

	address, err := invoke(stub, [][]byte{[]byte(write), testPayload})
	assert.Nil(t, err)
	assert.Equal(t, testPayloadAddress, string(address))

	payload, err = invoke(stub, [][]byte{[]byte(read), []byte(testPayloadAddress)})
	assert.Nil(t, err)
	assert.Equal(t, testPayload, payload)
}

func TestReadError(t *testing.T) {

	testErr := fmt.Errorf("read error")
	stub := prepareStub()
	stub.GetPrivateErr = testErr

	testPayload := getOperationBytes(getCreateOperation())
	payload, err := invoke(stub, [][]byte{[]byte(read), testPayload})
	assert.NotNil(t, err)
	assert.Nil(t, payload)
	assert.Contains(t, err.Error(), testErr.Error())
}

func TestRead_MissingAddress(t *testing.T) {

	stub := prepareStub()

	address, err := invoke(stub, [][]byte{[]byte(read), []byte("")})
	assert.NotNil(t, err)
	assert.Nil(t, address)
	assert.Contains(t, err.Error(), "missing content address")
}

func TestQuery(t *testing.T) {

	stub := prepareStub()

	testPayload := getOperationBytes(getCreateOperation())
	testPayloadAddress := encodedSHA256Hash(testPayload)

	address, err := invoke(stub, [][]byte{[]byte(write), testPayload})
	assert.Nil(t, err)
	assert.Equal(t, testPayloadAddress, string(address))

	payload, err := invoke(stub, [][]byte{[]byte(queryByID), []byte(testID)})
	assert.Nil(t, err)
	assert.NotNil(t, payload)
}

func TestQuery_MissingID(t *testing.T) {

	stub := prepareStub()

	payload, err := invoke(stub, [][]byte{[]byte(queryByID)})
	assert.NotNil(t, err)
	assert.Nil(t, payload)
	assert.Contains(t, err.Error(), "id is required")
}

func TestQuery_DocumentNotFound(t *testing.T) {

	stub := prepareStub()

	payload, err := invoke(stub, [][]byte{[]byte(queryByID), []byte(testID)})
	assert.NotNil(t, err)
	assert.Nil(t, payload)
	assert.Contains(t, err.Error(), "document not found")
}

func TestQueryError(t *testing.T) {

	testErr := fmt.Errorf("query error")
	stub := prepareStub()
	stub.GetPrivateQueryErr = testErr

	payload, err := invoke(stub, [][]byte{[]byte(queryByID), []byte(testID)})
	assert.NotNil(t, err)
	assert.Nil(t, payload)
	assert.Contains(t, err.Error(), testErr.Error())
}

func TestQuery_SortError(t *testing.T) {
	stub := prepareStub()

	testPayload := getOperationBytes(getCreateOperation())

	address, err := invoke(stub, [][]byte{[]byte(write), testPayload})
	assert.Nil(t, err)
	assert.NotNil(t, address)

	address, err = invoke(stub, [][]byte{[]byte(write), []byte("invalid json")})
	assert.Nil(t, err)
	assert.NotNil(t, address)

	payload, err := invoke(stub, [][]byte{[]byte(queryByID), []byte(testID)})
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "unexpected end of JSON input")
	assert.Nil(t, payload)
}

func TestSort(t *testing.T) {
	var operations [][]byte

	delete := &testOperation{ID: testID, Type: "delete", TransactionTime: 2, TransactionNumber: 1}
	update := &testOperation{ID: testID, Type: "update", TransactionTime: 1, TransactionNumber: 7}
	create := &testOperation{ID: testID, Type: "create", TransactionTime: 1, TransactionNumber: 1}

	operations = append(operations, getOperationBytes(delete))
	operations = append(operations, getOperationBytes(update))
	operations = append(operations, getOperationBytes(create))

	result, err := sortChronologically(operations)
	require.NoError(t, err)

	var first testOperation
	err = json.Unmarshal(result[0], &first)
	require.NoError(t, err)
	require.Equal(t, create.Type, first.Type)

	var second testOperation
	err = json.Unmarshal(result[1], &second)
	require.NoError(t, err)
	require.Equal(t, update.Type, second.Type)

	var third testOperation
	err = json.Unmarshal(result[2], &third)
	require.NoError(t, err)
	require.Equal(t, delete.Type, third.Type)
}

func TestSortError(t *testing.T) {
	var operations [][]byte
	operations = append(operations, []byte("invalid json"))
	operations = append(operations, []byte("invalid json"))

	result, err := sortChronologically(operations)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid character")
	require.Nil(t, result)
}

func TestWarmup(t *testing.T) {

	stub := prepareStub()

	payload, err := invoke(stub, [][]byte{[]byte(warmup)})
	assert.Nil(t, err)
	assert.Nil(t, payload)
}

func TestHandlePanic(t *testing.T) {

	stub := prepareStub()
	stub.GetPrivateErr = fmt.Errorf("panic")

	payload, err := invoke(stub, [][]byte{[]byte(read), []byte("address")})
	assert.NotNil(t, err)
	assert.Nil(t, payload)
	assert.Contains(t, err.Error(), "panic")
}

func encodedSHA256Hash(bytes []byte) string {

	h := crypto.SHA256.New()
	if _, err := h.Write(bytes); err != nil {
		panic(err)
	}

	return base64.URLEncoding.EncodeToString(h.Sum(nil))
}

func testInvalidFunctionName(t *testing.T, stub *mocks.MockStub) {

	// Test function name not provided
	_, err := invoke(stub, [][]byte{})
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "function name is required")

	// Test wrong function name provided
	_, err = invoke(stub, [][]byte{[]byte("test")})
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "invalid invoke function")

}

func prepareStub() *mocks.MockStub {
	return mocks.NewMockStub(ccName, New(ccName))
}

func checkInit(t *testing.T, stub *mocks.MockStub, args [][]byte) {
	txID := stub.GetTxID()
	if txID == "" {
		txID = "1"
	}
	res := stub.MockInit(txID, args)
	if res.Status != shim.OK {
		t.Fatalf("Init failed: %s", res.Message)
	}
}

func invoke(stub *mocks.MockStub, args [][]byte) ([]byte, error) {
	txID := stub.GetTxID()
	if txID == "" {
		txID = "1"
	}
	res := stub.MockInvoke(txID, args)
	if res.Status != shim.OK {
		return nil, fmt.Errorf("MockInvoke failed: %s", res.Message)
	}
	return res.Payload, nil
}

func getOperationBytes(op *testOperation) []byte {

	bytes, err := docutil.MarshalCanonical(op)
	if err != nil {
		panic(err)
	}

	return bytes
}

func getCreateOperation() *testOperation {
	return &testOperation{ID: testID, Type: "create", TransactionTime: 1, TransactionNumber: 1}
}

// testOperation  defines sample operation with smaallest subset of information
type testOperation struct {
	Type              string `json:"type"`
	ID                string `json:"id"`
	TransactionTime   uint64 `json:"transactionTime"`
	TransactionNumber uint64 `json:"transactionNumber"`
}
