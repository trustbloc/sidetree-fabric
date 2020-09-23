/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package txn

import (
	"crypto"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/stretchr/testify/require"
	"github.com/trustbloc/sidetree-core-go/pkg/api/protocol"

	cmdmocks "github.com/trustbloc/sidetree-fabric/cmd/chaincode/mocks"
	"github.com/trustbloc/sidetree-fabric/pkg/mocks"
	"github.com/trustbloc/sidetree-fabric/pkg/observer/common"
)

//go:generate counterfeiter -o ../mocks/protocolclientchannelprovider.gen.go --fake-name ProtocolClientChannelProvider . protocolClientChannelProvider

const (
	ccName = "sidetreetxncc"
	coll1  = "coll1"
)

func TestNew(t *testing.T) {
	req := require.New(t)

	pccp := &cmdmocks.ProtocolClientChannelProvider{}

	cc := New(ccName, pccp)
	req.NotNil(cc)

	req.Empty(cc.GetDBArtifacts([]string{coll1}))
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
	address, err := invoke(stub, [][]byte{[]byte(writeContent), []byte(coll1), testPayload})
	require.Nil(t, err)
	require.Equal(t, encodedSHA256Hash(testPayload), string(address))

	payload, err := invoke(stub, [][]byte{[]byte(readContent), []byte(coll1), []byte(address)})
	require.Nil(t, err)
	require.Equal(t, testPayload, payload)
}

func TestWriteError(t *testing.T) {

	testErr := fmt.Errorf("write error")
	stub := prepareStub()
	stub.PutPrivateErr = testErr

	payload, err := invoke(stub, [][]byte{[]byte(writeContent), []byte(coll1), []byte("address")})
	require.NotNil(t, err)
	require.Nil(t, payload)
	require.Contains(t, err.Error(), testErr.Error())
}

func TestWrite_MissingContent(t *testing.T) {

	stub := prepareStub()

	address, err := invoke(stub, [][]byte{[]byte(writeContent), []byte(coll1), []byte("")})
	require.NotNil(t, err)
	require.Nil(t, address)
	require.Contains(t, err.Error(), "collection and content are required")
}

func TestRead(t *testing.T) {

	stub := prepareStub()

	testPayload := []byte("Test")
	testPayloadAddress := encodedSHA256Hash(testPayload)

	payload, err := invoke(stub, [][]byte{[]byte(readContent), []byte(coll1), []byte(testPayloadAddress)})
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "not found")

	address, err := invoke(stub, [][]byte{[]byte(writeContent), []byte(coll1), testPayload})
	require.Nil(t, err)
	require.Equal(t, testPayloadAddress, string(address))

	payload, err = invoke(stub, [][]byte{[]byte(readContent), []byte(coll1), []byte(testPayloadAddress)})
	require.Nil(t, err)
	require.Equal(t, testPayload, payload)
}

func TestReadError(t *testing.T) {

	testErr := fmt.Errorf("read error")
	stub := prepareStub()
	stub.GetPrivateErr = testErr

	payload, err := invoke(stub, [][]byte{[]byte(readContent), []byte(coll1), []byte("address")})
	require.NotNil(t, err)
	require.Nil(t, payload)
	require.Contains(t, err.Error(), testErr.Error())
}

func TestRead_MissingAddress(t *testing.T) {

	stub := prepareStub()

	address, err := invoke(stub, [][]byte{[]byte(readContent), []byte(coll1), []byte("")})
	require.NotNil(t, err)
	require.Nil(t, address)
	require.Contains(t, err.Error(), "collection and content address are required")
}

func TestWriteAnchor(t *testing.T) {
	stub := prepareStubWithProtocol(protocol.Protocol{GenesisTime: 100})

	t.Run("Success", func(t *testing.T) {
		anchor := []byte("anchor1")
		txnInfoBytes := getTxnInfoBytes(t, 100)

		payload, err := invoke(stub, [][]byte{[]byte(writeAnchor), anchor, txnInfoBytes})
		require.NoError(t, err)
		require.Nil(t, payload)

		result, err := stub.GetState(common.AnchorPrefix + string(anchor))
		require.NoError(t, err)
		require.Equal(t, txnInfoBytes, result)
	})

	t.Run("Invalid txn info bytes", func(t *testing.T) {
		anchor := []byte("anchor2")

		payload, err := invoke(stub, [][]byte{[]byte(writeAnchor), anchor, []byte("invalid TxnInfo")})
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid transaction info payload")
		require.Nil(t, payload)
	})

	t.Run("Invalid protocol genesis time", func(t *testing.T) {
		anchor := []byte("anchor2")

		payload, err := invoke(stub, [][]byte{[]byte(writeAnchor), anchor, getTxnInfoBytes(t, 99)})
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid protocol genesis time in request")
		require.Nil(t, payload)

		result, err := stub.GetState(common.AnchorPrefix + string(anchor))
		require.NoError(t, err)
		require.Empty(t, result)
	})

	t.Run("Protocol client channel provider error", func(t *testing.T) {
		errExpected := fmt.Errorf("injected protocol client channel provider error")

		pccp := &cmdmocks.ProtocolClientChannelProvider{}
		pccp.ProtocolClientProviderForChannelReturns(nil, errExpected)

		stub := cmdmocks.NewMockStub(ccName, New(ccName, pccp))

		payload, err := invoke(stub, [][]byte{[]byte(writeAnchor), []byte("anchor"), getTxnInfoBytes(t, 101)})
		require.Error(t, err)
		require.Contains(t, err.Error(), errExpected.Error())
		require.Nil(t, payload)
	})

	t.Run("Protocol client provider error", func(t *testing.T) {
		errExpected := fmt.Errorf("injected protocol client provider error")

		pcp := &mocks.ProtocolClientProvider{}
		pcp.ForNamespaceReturns(nil, errExpected)

		pccp := &cmdmocks.ProtocolClientChannelProvider{}
		pccp.ProtocolClientProviderForChannelReturns(pcp, nil)

		stub := cmdmocks.NewMockStub(ccName, New(ccName, pccp))

		payload, err := invoke(stub, [][]byte{[]byte(writeAnchor), []byte("anchor"), getTxnInfoBytes(t, 100)})
		require.Error(t, err)
		require.Contains(t, err.Error(), errExpected.Error())
		require.Nil(t, payload)
	})

	t.Run("Protocol client error", func(t *testing.T) {
		errExpected := fmt.Errorf("injected protocol client error")

		pc := &mocks.ProtocolClient{}
		pc.GetReturns(nil, errExpected)

		pcp := &mocks.ProtocolClientProvider{}
		pcp.ForNamespaceReturns(pc, nil)

		pccp := &cmdmocks.ProtocolClientChannelProvider{}
		pccp.ProtocolClientProviderForChannelReturns(pcp, nil)

		stub := cmdmocks.NewMockStub(ccName, New(ccName, pccp))

		payload, err := invoke(stub, [][]byte{[]byte(writeAnchor), []byte("anchor"), getTxnInfoBytes(t, 100)})
		require.Error(t, err)
		require.Contains(t, err.Error(), errExpected.Error())
		require.Nil(t, payload)
	})
}

func getTxnInfoBytes(t *testing.T, protocolGenesisTime uint64) []byte {
	txnInfo := &common.TxnInfo{
		AnchorString:        "anchor",
		Namespace:           "ns",
		ProtocolGenesisTime: protocolGenesisTime,
	}

	txnInfoBytes, err := json.Marshal(txnInfo)
	require.NoError(t, err)

	return txnInfoBytes
}

func TestWriteAnchor_CheckRequiredArguments(t *testing.T) {
	stub := prepareStub()

	// missing args
	payload, err := invoke(stub, [][]byte{[]byte(writeAnchor)})
	require.NotNil(t, err)
	require.Nil(t, payload)
	require.Contains(t, err.Error(), "missing anchor string and/or txn info")

	// empty anchor
	payload, err = invoke(stub, [][]byte{[]byte(writeAnchor), []byte("")})
	require.NotNil(t, err)
	require.Nil(t, payload)
	require.Contains(t, err.Error(), "missing anchor string and/or txn info")

	// empty txn info
	payload, err = invoke(stub, [][]byte{[]byte(writeAnchor), []byte("address"), []byte("")})
	require.NotNil(t, err)
	require.Nil(t, payload)
	require.Contains(t, err.Error(), "missing anchor string and/or txn info")

	// suc
	payload, err = invoke(stub, [][]byte{[]byte(writeAnchor), []byte("address"), []byte("")})
	require.NotNil(t, err)
	require.Nil(t, payload)
	require.Contains(t, err.Error(), "missing anchor string and/or txn info")
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

	payload, err := invoke(stub, [][]byte{[]byte(readContent), []byte(coll1), []byte("address")})
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

func testInvalidFunctionName(t *testing.T, stub *cmdmocks.MockStub) {

	// Test function name not provided
	_, err := invoke(stub, [][]byte{})
	require.NotNil(t, err, "Function name is mandatory")

	// Test wrong function name provided
	_, err = invoke(stub, [][]byte{[]byte("test")})
	require.NotNil(t, err, "Should have failed due to wrong function name")

}

func prepareStub() *cmdmocks.MockStub {
	return prepareStubWithProtocol(protocol.Protocol{GenesisTime: 100})
}

func prepareStubWithProtocol(p protocol.Protocol) *cmdmocks.MockStub {
	pv := &mocks.ProtocolVersion{}
	pv.ProtocolReturns(p)

	pc := &mocks.ProtocolClient{}
	pc.GetReturns(pv, nil)

	pcp := &mocks.ProtocolClientProvider{}
	pcp.ForNamespaceReturns(pc, nil)

	pccp := &cmdmocks.ProtocolClientChannelProvider{}
	pccp.ProtocolClientProviderForChannelReturns(pcp, nil)

	return cmdmocks.NewMockStub(ccName, New(ccName, pccp))
}

func checkInit(t *testing.T, stub *cmdmocks.MockStub, args [][]byte) {
	txID := stub.GetTxID()
	if txID == "" {
		txID = "1"
	}
	res := stub.MockInit(txID, args)
	if res.Status != shim.OK {
		t.Fatalf("Init failed: %s", res.Message)
	}
}

func invoke(stub *cmdmocks.MockStub, args [][]byte) ([]byte, error) {
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
