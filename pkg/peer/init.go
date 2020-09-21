/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package peer

import (
	"strings"

	"github.com/hyperledger/fabric/common/flogging"
	ccapi "github.com/hyperledger/fabric/extensions/chaincode/api"
	logger "github.com/trustbloc/edge-core/pkg/log"
	"github.com/trustbloc/fabric-peer-ext/pkg/chaincode/ucc"
	extpeer "github.com/trustbloc/fabric-peer-ext/pkg/peer"
	"github.com/trustbloc/fabric-peer-ext/pkg/resource"
	"go.uber.org/zap"

	"github.com/trustbloc/sidetree-fabric/cmd/chaincode/doc"
	"github.com/trustbloc/sidetree-fabric/cmd/chaincode/file"
	"github.com/trustbloc/sidetree-fabric/cmd/chaincode/txn"
	"github.com/trustbloc/sidetree-fabric/pkg/client"
	"github.com/trustbloc/sidetree-fabric/pkg/context/operationqueue"
	"github.com/trustbloc/sidetree-fabric/pkg/peer/config"
	"github.com/trustbloc/sidetree-fabric/pkg/peer/discovery"
	"github.com/trustbloc/sidetree-fabric/pkg/peer/sidetreesvc"
	"github.com/trustbloc/sidetree-fabric/pkg/protocolversion"
	"github.com/trustbloc/sidetree-fabric/pkg/protocolversion/factoryregistry"
)

// Initialize initializes the required resources for peer startup
func Initialize() {
	logger.Initialize(&loggingProvider{})

	extpeer.Initialize()

	resource.Register(config.NewPeer)
	resource.Register(config.NewSidetreeProvider)
	resource.Register(client.NewBlockchainProvider)
	resource.Register(sidetreesvc.NewProvider)
	resource.Register(operationqueue.NewProvider)
	resource.Register(discovery.New)
	resource.Register(factoryregistry.New)

	// Register chaincode
	ucc.Register(func() ccapi.UserCC { return doc.New("document") })
	ucc.Register(func() ccapi.UserCC { return txn.New("sidetreetxn") })
	ucc.Register(func() ccapi.UserCC { return file.New("file") })

	protocolversion.RegisterFactories()
}

type loggingProvider struct {
}

func (lp *loggingProvider) GetLogger(module string) logger.Logger {
	return &fabricLogger{
		s: flogging.Global.ZapLogger(strings.ReplaceAll(module, "/", "_")).WithOptions(zap.AddCallerSkip(2)).Sugar(),
	}
}
