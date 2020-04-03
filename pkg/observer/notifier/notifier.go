/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package notifier

import (
	"strings"

	"github.com/hyperledger/fabric-protos-go/ledger/rwset/kvrwset"
	"github.com/hyperledger/fabric/common/flogging"
	gossipapi "github.com/hyperledger/fabric/extensions/gossip/api"
	sidetreeobserver "github.com/trustbloc/sidetree-core-go/pkg/observer"

	"github.com/trustbloc/sidetree-fabric/pkg/observer/common"
)

var logger = flogging.MustGetLogger("sidetree_observer")

// blockPublisher allows clients to add handlers for various block events
type blockPublisher interface {
	// AddWriteHandler adds a handler for KV writes
	AddWriteHandler(handler gossipapi.WriteHandler)
}

// Notifier receives anchor file 'write' events and sends them to a Go channel.
type Notifier struct {
	publisher             blockPublisher
	anchorFileAddressChan chan []sidetreeobserver.SidetreeTxn
}

// New return new instance of Notifier
func New(publisher blockPublisher) *Notifier {
	n := &Notifier{
		publisher:             publisher,
		anchorFileAddressChan: make(chan []sidetreeobserver.SidetreeTxn, 100),
	}

	n.publisher.AddWriteHandler(n.handleWrite)

	return n
}

// RegisterForSidetreeTxn register to get AnchorFileAddress value from writeset in the block committed by sidetreetxn_cc
func (n *Notifier) RegisterForSidetreeTxn() <-chan []sidetreeobserver.SidetreeTxn {
	return n.anchorFileAddressChan
}

func (n *Notifier) handleWrite(txMetadata gossipapi.TxMetadata, namespace string, kvWrite *kvrwset.KVWrite) error {
	if !kvWrite.IsDelete && strings.HasPrefix(kvWrite.Key, common.AnchorAddrPrefix) {
		logger.Debugf("found anchor address key[%s], value [%s]", kvWrite.Key, string(kvWrite.Value))
		n.anchorFileAddressChan <- []sidetreeobserver.SidetreeTxn{{TransactionTime: txMetadata.BlockNum, TransactionNumber: txMetadata.TxNum, AnchorAddress: string(kvWrite.Value)}}
	}
	return nil
}
