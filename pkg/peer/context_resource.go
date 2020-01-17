/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package peer

import (
	"fmt"

	txnapi "github.com/trustbloc/fabric-peer-ext/pkg/txn/api"
	sidetreectx "github.com/trustbloc/sidetree-fabric/pkg/context"
)

type contextResource struct {
	*sidetreectx.SidetreeContext
}

type txnServiceProvider interface {
	ForChannel(channelID string) (txnapi.Service, error)
}

func newContext(config sidetreeConfigProvider, txnProvider txnServiceProvider, dcasProvider dcasClientProvider) *contextResource {
	logger.Infof("Creating Sidetree context")

	ctx, err := sidetreectx.New(config, txnProvider, dcasProvider)
	if err != nil {
		panic(fmt.Sprintf("Error creating Sidetree context: %s", err))
	}

	return &contextResource{
		SidetreeContext: ctx,
	}
}
