/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package protocol

import (
	"fmt"
	"sort"

	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric/common/flogging"
	"github.com/trustbloc/sidetree-core-go/pkg/api/protocol"
)

var logger = flogging.MustGetLogger("protocol_client")

type blockchainInfoProvider interface {
	GetBlockchainInfo() (*common.BlockchainInfo, error)
}

// Client is a struct which holds a list of protocols.
type Client struct {
	protocols      []protocol.Version
	bcInfoProvider blockchainInfoProvider
}

//New initializes the protocol parameters from file
func New(protocolVersions []protocol.Version, bcInfoProvider blockchainInfoProvider) *Client {
	// Creating the list of the protocol versions
	var protocols []protocol.Version

	protocols = append(protocols, protocolVersions...)

	// Sorting the protocolParameter list based on blockChain start time
	sort.SliceStable(protocols, func(i, j int) bool {
		return protocols[j].Protocol().GenesisTime > protocols[i].Protocol().GenesisTime
	})

	return &Client{
		protocols:      protocols,
		bcInfoProvider: bcInfoProvider,
	}
}

//Current returns the latest version of protocol
func (c *Client) Current() (protocol.Version, error) {
	bcInfo, err := c.bcInfoProvider.GetBlockchainInfo()
	if err != nil {
		return nil, err
	}

	return c.Get(bcInfo.Height - 1)
}

// Get gets protocol version based on blockchain(transaction) time
func (c *Client) Get(transactionTime uint64) (protocol.Version, error) {
	logger.Debugf("available protocols: %s", c.protocols)

	for i := len(c.protocols) - 1; i >= 0; i-- {
		pv := c.protocols[i]
		p := pv.Protocol()

		logger.Debugf("Checking protocol for transaction time %d: %+v", transactionTime, p)

		if transactionTime >= p.GenesisTime {
			logger.Debugf("Found protocol for transaction time %d: %+v", transactionTime, p)

			return pv, nil
		}
	}

	return nil, fmt.Errorf("protocol parameters are not defined for blockchain time: %d", transactionTime)
}
