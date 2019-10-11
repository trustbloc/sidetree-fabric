/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package client

import (
	"github.com/bluele/gcache"
	"github.com/hyperledger/fabric/core/peer"
	"github.com/hyperledger/fabric/protos/common"
	"github.com/pkg/errors"
)

// Blockchain defines the functions of a Blockchain client
type Blockchain interface {
	GetBlockchainInfo() (*common.BlockchainInfo, error)
	GetBlockByNumber(blockNumber uint64) (*common.Block, error)
}

// BlockchainProvider manages multiple blockchain clients - one per channel
type BlockchainProvider struct {
	cache gcache.Cache
}

// NewBlockchainProvider returns a new blockchain client provider
func NewBlockchainProvider() *BlockchainProvider {
	return &BlockchainProvider{
		cache: gcache.New(0).LoaderFunc(func(channelID interface{}) (interface{}, error) {
			ledger := getLedger(channelID.(string))
			if ledger == nil {
				return nil, errors.Errorf("no ledger for channel [%s]", channelID)
			}
			return ledger, nil
		}).Build(),
	}
}

// ForChannel returns the blockchain client for the given channel
func (cp *BlockchainProvider) ForChannel(channelID string) (Blockchain, error) {
	client, err := cp.cache.Get(channelID)
	if err != nil {
		return nil, err
	}
	return client.(Blockchain), nil
}

var getLedger = func(channelID string) Blockchain {
	return peer.GetLedger(channelID)
}
