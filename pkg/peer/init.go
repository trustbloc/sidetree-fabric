/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package peer

import (
	ccapi "github.com/hyperledger/fabric/extensions/chaincode/api"

	"github.com/trustbloc/fabric-peer-ext/pkg/chaincode/ucc"
	"github.com/trustbloc/fabric-peer-ext/pkg/resource"

	"github.com/trustbloc/sidetree-fabric/cmd/chaincode/doc"
	"github.com/trustbloc/sidetree-fabric/cmd/chaincode/filescc"
	"github.com/trustbloc/sidetree-fabric/cmd/chaincode/txn"
	"github.com/trustbloc/sidetree-fabric/pkg/client"
	"github.com/trustbloc/sidetree-fabric/pkg/context/operationqueue"
	"github.com/trustbloc/sidetree-fabric/pkg/peer/config"
	"github.com/trustbloc/sidetree-fabric/pkg/peer/sidetreesvc"
)

// Initialize initializes the required resources for peer startup
func Initialize() {
	resource.Register(config.NewPeer)
	resource.Register(config.NewSidetreeProvider)
	resource.Register(client.NewBlockchainProvider)
	resource.Register(sidetreesvc.NewProvider)
	resource.Register(operationqueue.NewProvider)

	// Register chaincode
	ucc.Register(func() ccapi.UserCC { return doc.New("document_cc") })
	ucc.Register(func() ccapi.UserCC { return txn.New("sidetreetxn_cc") })
	ucc.Register(func() ccapi.UserCC { return filescc.New("files") })
}
