/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package blockchain

import (
	"github.com/pkg/errors"
	txnapi "github.com/trustbloc/fabric-peer-ext/pkg/txn/api"
	"github.com/trustbloc/sidetree-core-go/pkg/observer"

	"github.com/trustbloc/sidetree-fabric/pkg/common/transienterr"
)

const (
	writeAnchorFcn = "writeAnchor"
)

type txnServiceProvider interface {
	ForChannel(channelID string) (txnapi.Service, error)
}

// Client implements blockchain client for writing anchors
type Client struct {
	channelID     string
	chaincodeName string
	txnProvider   txnServiceProvider
}

// New returns a new blockchain client
func New(channelID, chaincodeName string, txnProvider txnServiceProvider) *Client {
	return &Client{
		channelID:     channelID,
		chaincodeName: chaincodeName,
		txnProvider:   txnProvider,
	}
}

// WriteAnchor writes anchor file address to blockchain
func (c *Client) WriteAnchor(anchor string) error {
	txnService, err := c.txnProvider.ForChannel(c.channelID)
	if err != nil {
		return err
	}

	_, err = txnService.EndorseAndCommit(&txnapi.Request{
		ChaincodeID: c.chaincodeName,
		Args:        [][]byte{[]byte(writeAnchorFcn), []byte(anchor)},
	})
	if err != nil {
		return transienterr.New(errors.Wrap(err, "failed to store anchor file address"), transienterr.CodeBlockchain)
	}

	return nil
}

// Read ledger transaction
func (c *Client) Read(sinceTransactionNumber int) (bool, *observer.SidetreeTxn) {
	// TODO: Not sure where/if this function is used
	panic("not implemented")
}
