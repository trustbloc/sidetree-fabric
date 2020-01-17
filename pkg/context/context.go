/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package context

import (
	"github.com/sirupsen/logrus"
	viper "github.com/spf13/viper2015"
	"github.com/trustbloc/fabric-peer-ext/pkg/collections/offledger/dcas/client"
	txnapi "github.com/trustbloc/fabric-peer-ext/pkg/txn/api"
	protocolApi "github.com/trustbloc/sidetree-core-go/pkg/api/protocol"
	"github.com/trustbloc/sidetree-core-go/pkg/batch"
	"github.com/trustbloc/sidetree-fabric/pkg/context/blockchain"
	"github.com/trustbloc/sidetree-fabric/pkg/context/cas"
	"github.com/trustbloc/sidetree-fabric/pkg/context/protocol"
)

const (
	keyProtocolFile     = "sidetree.protocol.file"
	defaultProtocolFile = "protocol.json"
)

var logger = logrus.New()

// SidetreeContext implements 'Fabric' version of Sidetree node context
type SidetreeContext struct {
	protocolClient   protocolApi.Client
	casClient        batch.CASClient
	blockchainClient batch.BlockchainClient
}

type txnServiceProvider interface {
	ForChannel(channelID string) (txnapi.Service, error)
}

type dcasClientProvider interface {
	ForChannel(channelID string) (client.DCAS, error)
}

type sidetreeConfigProvider interface {
	ChannelID() string
}

// New creates new Sidetree context
func New(cfg sidetreeConfigProvider, txnProvider txnServiceProvider, dcasProvider dcasClientProvider) (*SidetreeContext, error) {
	pc, err := getProtocolClient()
	if err != nil {
		logger.Errorf("Failed to load protocol: %s", err.Error())
		return nil, err
	}

	return &SidetreeContext{
		protocolClient:   pc,
		casClient:        cas.New(cfg.ChannelID(), dcasProvider),
		blockchainClient: blockchain.New(cfg.ChannelID(), txnProvider),
	}, nil
}

func getProtocolClient() (*protocol.Client, error) {
	protocolConfigFile := defaultProtocolFile
	if viper.IsSet(keyProtocolFile) {
		protocolConfigFile = viper.GetString(keyProtocolFile)
	}

	return protocol.New(protocolConfigFile)
}

// Protocol returns protocol client
func (m *SidetreeContext) Protocol() protocolApi.Client {
	return m.protocolClient
}

// Blockchain returns blockchain client
func (m *SidetreeContext) Blockchain() batch.BlockchainClient {
	return m.blockchainClient
}

// CAS returns content addressable storage client
func (m *SidetreeContext) CAS() batch.CASClient {
	return m.casClient
}
