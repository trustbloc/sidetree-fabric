/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package context

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	extmocks "github.com/trustbloc/fabric-peer-ext/pkg/mocks"
	"github.com/trustbloc/sidetree-core-go/pkg/batch/opqueue"

	"github.com/trustbloc/sidetree-fabric/pkg/config"
	ctxmocks "github.com/trustbloc/sidetree-fabric/pkg/context/mocks"
	"github.com/trustbloc/sidetree-fabric/pkg/mocks"
)

//go:generate counterfeiter -o ./../mocks/txnserviceprovider.gen.go --fake-name TxnServiceProvider . txnServiceProvider
//go:generate counterfeiter -o ./../mocks/txnservice.gen.go --fake-name TxnService github.com/trustbloc/fabric-peer-ext/pkg/txn/api.Service
//go:generate counterfeiter -o ./../mocks/opqueueprovider.gen.go --fake-name OperationQueueProvider . operationQueueProvider
//go:generate counterfeiter -o ./../mocks/casclient.gen.go --fake-name CasClient github.com/trustbloc/sidetree-core-go/pkg/api/cas.Client
//go:generate counterfeiter -o ./mocks/cachingopprocessorprovider.gen.go --fake-name CachingOpProcessorProvider . cachingOpProcessorProvider

const (
	channelID = "channel1"
	namespace = "did:sidetree"
	ccName    = "cc1"
	coll      = "coll1"
)

func TestNew(t *testing.T) {
	txnProvider := &mocks.TxnServiceProvider{}
	dcasProvider := &mocks.DCASClientProvider{}
	opQueueProvider := &mocks.OperationQueueProvider{}
	ledgerProvider := &extmocks.LedgerProvider{}
	cacheUpdater := &ctxmocks.CachingOpProcessorProvider{}

	errExpected := errors.New("injected op queue error")
	opQueueProvider.CreateReturns(nil, errExpected)

	dcasCfg := config.DCAS{
		ChaincodeName: ccName,
		Collection:    coll,
	}

	p := &Providers{
		TxnProvider:                txnProvider,
		DCASProvider:               dcasProvider,
		OperationQueueProvider:     opQueueProvider,
		LedgerProvider:             ledgerProvider,
		OperationProcessorProvider: cacheUpdater,
	}

	casClient := &mocks.CasClient{}

	sctx, err := New(channelID, namespace, dcasCfg, casClient, nil, p)
	require.EqualError(t, err, errExpected.Error())
	require.Nil(t, sctx)

	opQueueProvider.CreateReturns(&opqueue.MemQueue{}, nil)

	sctx, err = New(channelID, namespace, dcasCfg, casClient, nil, p)
	require.NoError(t, err)
	require.NotNil(t, sctx)

	require.NotNil(t, sctx.Protocol())
	require.NotNil(t, sctx.CAS())
	require.NotNil(t, sctx.Anchor())
	require.NotEmpty(t, sctx.Namespace())
	require.NotNil(t, sctx.OperationQueue())
}
