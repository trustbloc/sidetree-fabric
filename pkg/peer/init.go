/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package peer

import (
	ccapi "github.com/hyperledger/fabric/extensions/chaincode/api"
	"github.com/trustbloc/fabric-peer-ext/pkg/chaincode/ucc"
	extpeer "github.com/trustbloc/fabric-peer-ext/pkg/peer"
	"github.com/trustbloc/fabric-peer-ext/pkg/resource"

	"github.com/trustbloc/sidetree-fabric/cmd/chaincode/doc"
	"github.com/trustbloc/sidetree-fabric/cmd/chaincode/file"
	"github.com/trustbloc/sidetree-fabric/cmd/chaincode/txn"
	"github.com/trustbloc/sidetree-fabric/pkg/client"
	"github.com/trustbloc/sidetree-fabric/pkg/context/operationqueue"
	"github.com/trustbloc/sidetree-fabric/pkg/peer/config"
	"github.com/trustbloc/sidetree-fabric/pkg/peer/discovery"
	"github.com/trustbloc/sidetree-fabric/pkg/peer/sidetreesvc"
)

// Initialize initializes the required resources for peer startup
func Initialize() {
	extpeer.Initialize()

	resource.Register(config.NewPeer)
	resource.Register(config.NewSidetreeProvider)
	resource.Register(client.NewBlockchainProvider)
	resource.Register(sidetreesvc.NewProvider)
	resource.Register(operationqueue.NewProvider)
	resource.Register(discovery.New)

	// Register chaincode
	ucc.Register(func() ccapi.UserCC { return doc.New("document") })
	ucc.Register(func() ccapi.UserCC { return txn.New("sidetreetxn") })
	ucc.Register(func() ccapi.UserCC { return file.New("file") })
}
