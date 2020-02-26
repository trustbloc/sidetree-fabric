/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package doc

import (
	"fmt"
	"testing"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/stretchr/testify/require"
	"github.com/trustbloc/sidetree-fabric/cmd/chaincode/mocks"
)

const ccName = "document_cc"

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

	_, err := invoke(stub, [][]byte{})
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "invoke is not supported")

	// run last
	checkInit(t, stub, [][]byte{})
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
