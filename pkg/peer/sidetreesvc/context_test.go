/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package sidetreesvc

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	protocolApi "github.com/trustbloc/sidetree-core-go/pkg/api/protocol"
	"github.com/trustbloc/sidetree-fabric/pkg/peer/config"
	"github.com/trustbloc/sidetree-fabric/pkg/peer/mocks"
)

//go:generate counterfeiter -o ../mocks/txnserviceprovider.gen.go --fake-name TxnServiceProvider . txnServiceProvider
//go:generate counterfeiter -o ../mocks/dcasprovider.gen.go --fake-name DCASClientProvider . dcasClientProvider

func TestContext(t *testing.T) {
	nsCfg := config.Namespace{
		Namespace: didTrustblocNamespace,
		BasePath:  didTrustblocBasePath,
	}

	txnProvider := &mocks.TxnServiceProvider{}
	dcasProvider := &mocks.DCASClientProvider{}

	t.Run("Success", func(t *testing.T) {
		protocolVersions := map[string]protocolApi.Protocol{
			"0.5": {
				StartingBlockChainTime:       100,
				HashAlgorithmInMultiHashCode: 18,
				MaxOperationsPerBatch:        100,
				MaxOperationByteSize:         1000,
			},
		}

		stConfigService := &mocks.SidetreeConfigService{}
		stConfigService.LoadProtocolsReturns(protocolVersions, nil)

		ctx, err := newContext(channel1, nsCfg, stConfigService, txnProvider, dcasProvider)
		require.NoError(t, err)
		require.NotNil(t, ctx)

		require.NotNil(t, ctx.BatchWriter())

		require.NoError(t, ctx.Start())

		time.Sleep(20 * time.Millisecond)

		ctx.Stop()
	})

	t.Run("No protocols -> error", func(t *testing.T) {
		stConfigService := &mocks.SidetreeConfigService{}

		ctx, err := newContext(channel1, nsCfg, stConfigService, txnProvider, dcasProvider)
		require.Error(t, err)
		require.Contains(t, err.Error(), "no protocols defined")
		require.Nil(t, ctx)
	})

	t.Run("Load protocols -> error", func(t *testing.T) {
		errExpected := errors.New("injected sidetreeCfgService error")
		stConfigService := &mocks.SidetreeConfigService{}
		stConfigService.LoadProtocolsReturns(nil, errExpected)

		ctx, err := newContext(channel1, nsCfg, stConfigService, txnProvider, dcasProvider)
		require.EqualError(t, err, errExpected.Error())
		require.Nil(t, ctx)
	})
}
