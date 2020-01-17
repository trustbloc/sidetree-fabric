/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package context

import (
	"testing"

	viper "github.com/spf13/viper2015"
	"github.com/stretchr/testify/require"
	"github.com/trustbloc/sidetree-fabric/pkg/mocks"
)

//go:generate counterfeiter -o ./../mocks/txnserviceprovider.gen.go --fake-name TxnServiceProvider . txnServiceProvider
//go:generate counterfeiter -o ./../mocks/txnservice.gen.go --fake-name TxnService github.com/trustbloc/fabric-peer-ext/pkg/txn/api.Service

const (
	sdkConfigFile      = "./testdata/config.yaml"
	protocolConfigFile = "./testdata/protocol.json"
)

func TestNew(t *testing.T) {
	viper.Set(keyProtocolFile, protocolConfigFile)

	txnProvider := &mocks.TxnServiceProvider{}
	dcasProvider := &mocks.DCASClientProvider{}

	sctx, err := New(&configProvider{}, txnProvider, dcasProvider)
	require.Nil(t, err)

	require.NotNil(t, sctx.Protocol())
	require.NotNil(t, sctx.CAS())
	require.NotNil(t, sctx.Blockchain())

}

func TestNew_ProtocolError(t *testing.T) {
	viper.Set(keyProtocolFile, "./invalid/protocol.json")

	txnProvider := &mocks.TxnServiceProvider{}
	dcasProvider := &mocks.DCASClientProvider{}

	sctx, err := New(&configProvider{}, txnProvider, dcasProvider)
	require.NotNil(t, err)
	require.Nil(t, sctx)
	require.Contains(t, err.Error(), "no such file or directory")

}

type configProvider struct {
	channelID string
}

func (p *configProvider) ChannelID() string {
	return p.channelID
}
