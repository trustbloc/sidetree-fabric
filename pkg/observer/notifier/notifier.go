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

var logger = flogging.MustGetLogger("notifier")

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
		if namespace != common.SidetreeNs {
			logger.Debugf("write NameSpace: %s not equal %s will skip this kvrwset", namespace, common.SidetreeNs)
			return nil
		}
		if !kvWrite.IsDelete && strings.HasPrefix(kvWrite.Key, common.AnchorAddrPrefix) {
			logger.Debugf("found anchor address key[%s], value [%s]", kvWrite.Key, string(kvWrite.Value))
			anchorFileAddressChan <- []sidetreeobserver.SidetreeTxn{{TransactionTime: txMetadata.BlockNum, TransactionNumber: txMetadata.TxNum, AnchorAddress: string(kvWrite.Value)}}
		}
		return nil
	})

	return anchorFileAddressChan
}
