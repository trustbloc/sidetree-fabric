/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package client

import (
	"github.com/bluele/gcache"
	cb "github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric/common/flogging"
	"github.com/pkg/errors"
	"github.com/trustbloc/fabric-peer-ext/pkg/collections/common"
)

var logger = flogging.MustGetLogger("sidetree_client")

// ErrNoLedger indicates that the ledger (channel) doesn't exist
var ErrNoLedger = errors.New("no ledger")

// Blockchain defines the functions of a Blockchain client
type Blockchain interface {
	GetBlockchainInfo() (*cb.BlockchainInfo, error)
	GetBlockByNumber(blockNumber uint64) (*cb.Block, error)
	GetBlockByHash(blockHash []byte) (*cb.Block, error)
}

// BlockchainProvider manages multiple blockchain clients - one per channel
type BlockchainProvider struct {
	cache gcache.Cache
}

// NewBlockchainProvider returns a new blockchain client provider
func NewBlockchainProvider(ledgerProvider common.LedgerProvider) *BlockchainProvider {
	logger.Infof("Creating blockchain provider")

	return &BlockchainProvider{
		cache: gcache.New(0).LoaderFunc(func(channelID interface{}) (interface{}, error) {
			ledger := ledgerProvider.GetLedger(channelID.(string))
			if ledger == nil {
				return nil, ErrNoLedger
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
