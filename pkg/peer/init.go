/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package peer

import (
	"github.com/hyperledger/fabric/common/flogging"
	ccapi "github.com/hyperledger/fabric/extensions/chaincode/api"
	"github.com/trustbloc/fabric-peer-ext/pkg/chaincode/ucc"
	dcas "github.com/trustbloc/fabric-peer-ext/pkg/collections/offledger/dcas/client"
	"github.com/trustbloc/fabric-peer-ext/pkg/resource"
	"github.com/trustbloc/sidetree-fabric/cmd/chaincode/doc"
	"github.com/trustbloc/sidetree-fabric/cmd/chaincode/txn"
	"github.com/trustbloc/sidetree-fabric/pkg/client"
	"github.com/trustbloc/sidetree-fabric/pkg/context/store"
	"github.com/trustbloc/sidetree-fabric/pkg/role"
)

var logger = flogging.MustGetLogger("sidetree_peer")

const (
	didDocNamespace   = "did:sidetree"
	keyConfigFile     = "sidetree.config.file"
	defaultConfigFile = "config.yaml"
)

// Initialize initializes the required resources for peer startup
func Initialize() {
	resource.Register(client.NewBlockchainProvider)
	resource.Register(newSidetreeConfigResource)
	resource.Register(newStore)
	resource.Register(newContext)
	resource.Register(newObserver)
	resource.Register(newBatchWriter)

	if role.IsResolver() || role.IsBatchWriter() {
		resource.Register(newRESTService)
	}

	// Register chaincode
	ucc.Register(func() ccapi.UserCC { return doc.New("document_cc") })
	ucc.Register(func() ccapi.UserCC { return txn.New("sidetreetxn_cc") })
}

type sidetreeConfigProvider interface {
	ChannelID() string
}

type dcasClientProvider interface {
	ForChannel(channelID string) (dcas.DCAS, error)
}

func newStore(config sidetreeConfigProvider, storeProvider dcasClientProvider) *store.Client {
	logger.Infof("Creating operation store client")
	return store.New(config.ChannelID(), didDocNamespace, storeProvider)
}
