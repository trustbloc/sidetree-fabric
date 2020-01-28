/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package context

import (
	"testing"

	"github.com/stretchr/testify/require"
	protocolApi "github.com/trustbloc/sidetree-core-go/pkg/api/protocol"
	"github.com/trustbloc/sidetree-fabric/pkg/mocks"
)

//go:generate counterfeiter -o ./../mocks/txnserviceprovider.gen.go --fake-name TxnServiceProvider . txnServiceProvider
//go:generate counterfeiter -o ./../mocks/txnservice.gen.go --fake-name TxnService github.com/trustbloc/fabric-peer-ext/pkg/txn/api.Service

const (
	channelID = "channel1"
	namespace = "did:sidetree"
)

func TestNew(t *testing.T) {
	txnProvider := &mocks.TxnServiceProvider{}
	dcasProvider := &mocks.DCASClientProvider{}

	protocolVersions := map[string]protocolApi.Protocol{}

	sctx := New(channelID, namespace, protocolVersions, txnProvider, dcasProvider)

	require.NotNil(t, sctx.Protocol())
	require.NotNil(t, sctx.CAS())
	require.NotNil(t, sctx.Blockchain())
	require.NotEmpty(t, sctx.Namespace())
}
