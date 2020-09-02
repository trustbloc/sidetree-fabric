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
	protocols      []protocol.Protocol
	bcInfoProvider blockchainInfoProvider
}

//New initializes the protocol parameters from file
func New(protocolVersions map[string]protocol.Protocol, bcInfoProvider blockchainInfoProvider) *Client {
	// Creating the list of the protocol versions
	protocols := make([]protocol.Protocol, 0, len(protocolVersions))
	for _, v := range protocolVersions {
		protocols = append(protocols, v)
	}

	// Sorting the protocolParameter list based on blockChain start time
	sort.SliceStable(protocols, func(i, j int) bool {
		return protocols[j].GenesisTime > protocols[i].GenesisTime
	})

	return &Client{
		protocols:      protocols,
		bcInfoProvider: bcInfoProvider,
	}
}

//Current returns the latest version of protocol
func (c *Client) Current() (protocol.Protocol, error) {
	bcInfo, err := c.bcInfoProvider.GetBlockchainInfo()
	if err != nil {
		return protocol.Protocol{}, err
	}

	return c.Get(bcInfo.Height - 1)
}

// Get gets protocol version based on blockchain(transaction) time
func (c *Client) Get(transactionTime uint64) (protocol.Protocol, error) {
	logger.Debugf("available protocols: %v", c.protocols)

	for i := len(c.protocols) - 1; i >= 0; i-- {
		if transactionTime >= c.protocols[i].GenesisTime {
			return c.protocols[i], nil
		}
	}

	return protocol.Protocol{}, fmt.Errorf("protocol parameters are not defined for blockchain time: %d", transactionTime)
}
