/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package txn

import (
	"crypto"
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/trustbloc/sidetree-fabric/cmd/chaincode/mocks"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/stretchr/testify/require"
)

const (
	ccName = "sidetreetxncc"
)

func TestNew(t *testing.T) {
	req := require.New(t)

	cc := New(ccName)
	req.NotNil(cc)

	req.Nil(cc.GetDBArtifacts())
	req.Equal(ccName, cc.Name())
	req.Equal(ccVersion, cc.Version())
	req.Equal(cc, cc.Chaincode())
}

func TestInvoke(t *testing.T) {

	stub := prepareStub()

	testInvalidFunctionName(t, stub)

	// Run last
	checkInit(t, stub, [][]byte{})
}

func TestWrite(t *testing.T) {

	stub := prepareStub()

	testPayload := []byte("Test")
	address, err := invoke(stub, [][]byte{[]byte(writeContent), testPayload})
	require.Nil(t, err)
	require.Equal(t, encodedSHA256Hash(testPayload), string(address))

	payload, err := invoke(stub, [][]byte{[]byte(readContent), []byte(address)})
	require.Nil(t, err)
	require.Equal(t, testPayload, payload)
}

func TestWriteError(t *testing.T) {

	testErr := fmt.Errorf("write error")
	stub := prepareStub()
	stub.PutPrivateErr = testErr

	payload, err := invoke(stub, [][]byte{[]byte(writeContent), []byte("address")})
	require.NotNil(t, err)
	require.Nil(t, payload)
	require.Contains(t, err.Error(), testErr.Error())
}

func TestWrite_MissingContent(t *testing.T) {

	stub := prepareStub()

	address, err := invoke(stub, [][]byte{[]byte(writeContent), []byte("")})
	require.NotNil(t, err)
	require.Nil(t, address)
	require.Contains(t, err.Error(), "missing content")
}

func TestRead(t *testing.T) {

	stub := prepareStub()

	testPayload := []byte("Test")
	testPayloadAddress := encodedSHA256Hash(testPayload)

	payload, err := invoke(stub, [][]byte{[]byte(readContent), []byte(testPayloadAddress)})
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "not found")

	address, err := invoke(stub, [][]byte{[]byte(writeContent), testPayload})
	require.Nil(t, err)
	require.Equal(t, testPayloadAddress, string(address))

	payload, err = invoke(stub, [][]byte{[]byte(readContent), []byte(testPayloadAddress)})
	require.Nil(t, err)
	require.Equal(t, testPayload, payload)
}

func TestReadError(t *testing.T) {

	testErr := fmt.Errorf("read error")
	stub := prepareStub()
	stub.GetPrivateErr = testErr

	payload, err := invoke(stub, [][]byte{[]byte(readContent), []byte("address")})
	require.NotNil(t, err)
	require.Nil(t, payload)
	require.Contains(t, err.Error(), testErr.Error())
}

func TestRead_MissingAddress(t *testing.T) {

	stub := prepareStub()

	address, err := invoke(stub, [][]byte{[]byte(readContent), []byte("")})
	require.NotNil(t, err)
	require.Nil(t, address)
	require.Contains(t, err.Error(), "missing content address")
}

func TestWriteAnchor(t *testing.T) {

	stub := prepareStub()

	anchorAddress := []byte("Addr")
	payload, err := invoke(stub, [][]byte{[]byte(writeAnchor), anchorAddress})
	require.Nil(t, err)
	require.Nil(t, payload)

	result, err := stub.GetState(anchorAddrPrefix + string(anchorAddress))
	require.Nil(t, err)
	require.Equal(t, anchorAddress, result)
}

func TestWriteAnchor_MissingAnchorAddress(t *testing.T) {

	stub := prepareStub()

	payload, err := invoke(stub, [][]byte{[]byte(writeAnchor), []byte("")})
	require.NotNil(t, err)
	require.Nil(t, payload)
	require.Contains(t, err.Error(), "missing anchor file address")
}

func TestAnchorBatch(t *testing.T) {

	stub := prepareStub()

	batch := []byte("Ops")
	anchor := []byte("anchor")
	payload, err := invoke(stub, [][]byte{[]byte(anchorBatch), batch, anchor})
	require.Nil(t, err)
	require.Nil(t, payload)

	result, err := stub.GetState(anchorAddrPrefix + encodedSHA256Hash(anchor))
	require.Nil(t, err)
	require.Equal(t, string(result), encodedSHA256Hash(anchor))

}

func TestAnchorBatch_CASClientError(t *testing.T) {

	stub := prepareStub()
	stub.PutPrivateErr = fmt.Errorf("write error")

	payload, err := invoke(stub, [][]byte{[]byte(anchorBatch), []byte("Ops"), []byte("anchor")})
	require.NotNil(t, err)
	require.Nil(t, payload)
	require.Contains(t, err.Error(), "write error")
}

func TestAnchorBatch_PutStateError(t *testing.T) {

	stub := prepareStub()
	stub.MockStub.TxID = ""

	res := stub.MockInvoke("", [][]byte{[]byte(anchorBatch), []byte("Ops"), []byte("anchor")})
	require.NotEqual(t, res.Status, shim.OK)
}

func TestAnchorBatch_MissingRequiredParameters(t *testing.T) {

	stub := prepareStub()

	payload, err := invoke(stub, [][]byte{[]byte(anchorBatch), []byte("")})
	require.NotNil(t, err)
	require.Nil(t, payload)
	require.Contains(t, err.Error(), "batch and anchor files are required")

	payload, err = invoke(stub, [][]byte{[]byte(anchorBatch), []byte("Ops"), []byte("")})
	require.NotNil(t, err)
	require.Nil(t, payload)
	require.Contains(t, err.Error(), "batch and anchor files are required")
}

func TestWarmup(t *testing.T) {

	stub := prepareStub()

	payload, err := invoke(stub, [][]byte{[]byte(warmup)})
	require.Nil(t, err)
	require.Nil(t, payload)
}

func TestHandlePanic(t *testing.T) {

	stub := prepareStub()
	stub.GetPrivateErr = fmt.Errorf("panic")

	payload, err := invoke(stub, [][]byte{[]byte(readContent), []byte("address")})
	require.NotNil(t, err)
	require.Nil(t, payload)
	require.Contains(t, err.Error(), "panic")
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
	require.NotNil(t, err, "Function name is mandatory")

	// Test wrong function name provided
	_, err = invoke(stub, [][]byte{[]byte("test")})
	require.NotNil(t, err, "Should have failed due to wrong function name")

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
