/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package notifier

import (
	"strings"

	"github.com/hyperledger/fabric/common/flogging"
	gossipapi "github.com/hyperledger/fabric/extensions/gossip/api"
	"github.com/hyperledger/fabric/protos/ledger/rwset/kvrwset"
	sidetreeobserver "github.com/trustbloc/sidetree-core-go/pkg/observer"
)

var logger = flogging.MustGetLogger("notifier")

const (
	sideTreeTxnCCName = "sidetreetxn_cc"
	// anchor address prefix
	anchorAddrPrefix = "sidetreetxn_"
)

// blockPublisher allows clients to add handlers for various block events
type blockPublisher interface {
	// AddWriteHandler adds a handler for KV writes
	AddWriteHandler(handler gossipapi.WriteHandler)
}

// Notifier holds the gossip adapter and channel id
type Notifier struct {
	publisher blockPublisher
}

// New return new instance of Notifier
func New(publisher blockPublisher) *Notifier {
	return &Notifier{publisher: publisher}
}

// RegisterForSidetreeTxn register to get AnchorFileAddress value from writeset in the block committed by sidetreetxn_cc
func (n *Notifier) RegisterForSidetreeTxn() <-chan []sidetreeobserver.SidetreeTxn {
	anchorFileAddressChan := make(chan []sidetreeobserver.SidetreeTxn, 100)
	n.publisher.AddWriteHandler(func(txMetadata gossipapi.TxMetadata, namespace string, kvWrite *kvrwset.KVWrite) error {
		if namespace != sideTreeTxnCCName {
			logger.Debugf("write NameSpace: %s not equal %s will skip this kvrwset", namespace, sideTreeTxnCCName)
			return nil
		}
		if !kvWrite.IsDelete && strings.HasPrefix(kvWrite.Key, anchorAddrPrefix) {
			logger.Debugf("found anchor address key[%s], value [%s]", kvWrite.Key, string(kvWrite.Value))
			anchorFileAddressChan <- []sidetreeobserver.SidetreeTxn{{TransactionTime: txMetadata.BlockNum, TransactionNumber: txMetadata.TxNum, AnchorAddress: string(kvWrite.Value)}}
		}
		return nil
	})

	return anchorFileAddressChan
}
