/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package context

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/trustbloc/sidetree-core-go/pkg/mocks"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/stretchr/testify/require"

	fabMocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
)

const (
	sdkConfigFile      = "./testdata/config.yaml"
	protocolConfigFile = "./testdata/protocol.json"
)

func TestNew(t *testing.T) {
	config := viper.New()

	config.Set(keyConfigFile, sdkConfigFile)
	config.Set(keyProtocolFile, protocolConfigFile)

	sctx, err := New(config)
	require.Nil(t, err)

	require.NotNil(t, sctx.Protocol())
	require.NotNil(t, sctx.CAS())
	require.NotNil(t, sctx.Blockchain())
	require.NotNil(t, sctx.OperationStore())

}

func TestNewSDKConfigError(t *testing.T) {
	config := viper.New()

	config.Set(keyConfigFile, "./invalid/config.yaml")
	config.Set(keyProtocolFile, protocolConfigFile)

	sctx, err := New(config)
	require.NotNil(t, err)
	require.Nil(t, sctx)
	require.Contains(t, err.Error(), "failed to initialize configuration")

}

func TestSidetreeConfigError(t *testing.T) {
	config := viper.New()

	config.Set(keyConfigFile, "./testdata/config-nosidetree.yaml")
	config.Set(keyProtocolFile, protocolConfigFile)

	sctx, err := New(config)
	require.NotNil(t, err)
	require.Nil(t, sctx)
	require.Contains(t, err.Error(), "sidetree configuration key not found")

}

func TestNewProtocolError(t *testing.T) {
	config := viper.New()

	config.Set(keyConfigFile, sdkConfigFile)
	config.Set(keyProtocolFile, "./invalid/protocol.json")

	sctx, err := New(config)
	require.NotNil(t, err)
	require.Nil(t, sctx)
	require.Contains(t, err.Error(), "no such file or directory")

}

func TestNewSidetreeContext(t *testing.T) {
	ctx := mockChannelProvider("mychannel")
	sctx, err := newSidetreeContext(ctx, mocks.NewMockProtocolClient())
	require.Nil(t, err)
	require.NotNil(t, sctx)

	require.NotNil(t, sctx.Protocol())
	require.NotNil(t, sctx.CAS())
	require.NotNil(t, sctx.Blockchain())
	require.NotNil(t, sctx.OperationStore())

}

func mockChannelProvider(channelID string) context.ChannelProvider {
	channelProvider := func() (context.Channel, error) {
		return fabMocks.NewMockChannel(channelID)
	}
	return channelProvider
}
