/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package filescc

import (
	"testing"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/stretchr/testify/require"
)

const (
	ccName = "files_cc"
)

func TestFilesCC(t *testing.T) {
	cc := New(ccName)

	require.NotNil(t, cc)
	require.Equal(t, ccName, cc.Name())
	require.Equal(t, v1, cc.Version())
	require.True(t, cc.Chaincode() == cc)
	require.Nil(t, cc.GetDBArtifacts())

	resp := cc.Init(nil)
	require.NotNil(t, resp)
	require.Equal(t, int32(shim.OK), resp.Status)

	resp = cc.Invoke(nil)
	require.NotNil(t, resp)
	require.Equal(t, int32(shim.ERROR), resp.Status)
}
