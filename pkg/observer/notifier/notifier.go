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

const defaultNotificationChannelBufferSize = 100

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
	channelID             string
	publisher             blockPublisher
	anchorFileAddressChan chan []sidetreeobserver.SidetreeTxn
}

// New return new instance of Notifier
func New(channelID string, bpProvider blockPublisherProvider) *Notifier {
	n := &Notifier{
		channelID:             channelID,
		publisher:             bpProvider.ForChannel(channelID),
		anchorFileAddressChan: make(chan []sidetreeobserver.SidetreeTxn, defaultNotificationChannelBufferSize),
	}

	n.publisher.AddWriteHandler(n.handleWrite)

	logger.Infof("[%s] Started notifier", channelID)

	return n
}

// RegisterForSidetreeTxn register to get AnchorFileAddress value from writeset in the block committed by sidetreetxn_cc
func (n *Notifier) RegisterForSidetreeTxn() <-chan []sidetreeobserver.SidetreeTxn {
	return n.anchorFileAddressChan
}

func (n *Notifier) handleWrite(txMetadata gossipapi.TxMetadata, namespace string, kvWrite *kvrwset.KVWrite) error {
	if kvWrite.IsDelete || !strings.HasPrefix(kvWrite.Key, common.AnchorAddrPrefix) {
		return nil
	}

	logger.Debugf("[%s] Found anchor address key[%s], value [%s]", n.channelID, kvWrite.Key, string(kvWrite.Value))

	// If the channel buffer gets full, reject the event since blocking the BlockPublisher can have serious consequences.
	// If an event is rejected, the monitor should process it at a later time.

	select {
	case n.anchorFileAddressChan <- []sidetreeobserver.SidetreeTxn{{TransactionTime: txMetadata.BlockNum, TransactionNumber: txMetadata.TxNum, AnchorAddress: string(kvWrite.Value)}}:
		logger.Debugf("[%s] Submitted anchor address key[%s], value [%s]", n.channelID, kvWrite.Key, string(kvWrite.Value))
	default:
		logger.Warnf("[%s] Could not submit anchor address key[%s], value [%s]", n.channelID, kvWrite.Key, string(kvWrite.Value))
	}

	return nil
}
