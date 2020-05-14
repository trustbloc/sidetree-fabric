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

	"github.com/trustbloc/sidetree-fabric/pkg/observer/common"
)

var logger = flogging.MustGetLogger("sidetree_observer")

// blockPublisher allows clients to add handlers for various block events
type blockPublisher interface {
	// AddWriteHandler adds a handler for KV writes
	AddWriteHandler(handler gossipapi.WriteHandler)
}

type blockPublisherProvider interface {
	ForChannel(channelID string) gossipapi.BlockPublisher
}

// Notifier receives anchor file 'write' events and sends them to a Go channel.
type Notifier struct {
	channelID string
	publisher blockPublisher
	txnChan   chan<- gossipapi.TxMetadata
}

// New return new instance of Notifier
func New(channelID string, bpProvider blockPublisherProvider, txnChan chan<- gossipapi.TxMetadata) *Notifier {
	n := &Notifier{
		channelID: channelID,
		publisher: bpProvider.ForChannel(channelID),
		txnChan:   txnChan,
	}

	n.publisher.AddWriteHandler(n.handleWrite)

	logger.Infof("[%s] Started notifier", channelID)

	return n
}

func (n *Notifier) handleWrite(txMetadata gossipapi.TxMetadata, namespace string, kvWrite *kvrwset.KVWrite) error {
	if kvWrite.IsDelete || !strings.HasPrefix(kvWrite.Key, common.AnchorAddrPrefix) {
		return nil
	}

	logger.Debugf("[%s] Found anchor address key[%s], value [%s] in transaction - Block %d, TxnNum %d", n.channelID, kvWrite.Key, string(kvWrite.Value), txMetadata.BlockNum, txMetadata.TxNum)

	// If the channel buffer gets full, reject the event since blocking the BlockPublisher can have serious consequences.
	// No worries, since the block will be processed at a later time.

	select {
	case n.txnChan <- txMetadata:
		logger.Debugf("[%s] Submitted notification about anchor write in transaction - Block %d, TxnNum %d", n.channelID, txMetadata.BlockNum, txMetadata.TxNum)
	default:
		logger.Infof("[%s] Unable to submit notification about anchor write in transaction - Block %d, TxnNum %d", n.channelID, txMetadata.BlockNum, txMetadata.TxNum)
	}

	return nil
}
