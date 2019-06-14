/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package notifier

import (
	"testing"
	"time"

	gossipapi "github.com/hyperledger/fabric/extensions/gossip/api"
	"github.com/hyperledger/fabric/protos/ledger/rwset/kvrwset"
	"github.com/stretchr/testify/require"
)

const (
	testChannel = "testChannel"
	k1          = "key1"
	v1          = "value1"
)

func TestRegisterForAnchorFileAddress(t *testing.T) {
	p := &mockBlockPublisher{}
	notifier := New(p)

	t.Run("test key in kvrwset is deleted", func(t *testing.T) {
		// register to receive sidetree txn value
		sideTreeTxnCh := notifier.RegisterForSidetreeTxn()
		done := make(chan SideTreeTxn, 1)
		go func() {
			for {
				select {
				case sideTreeTxn := <-sideTreeTxnCh:
					done <- sideTreeTxn
				case <-time.After(1 * time.Second):
					done <- SideTreeTxn{}
				}
			}
		}()
		require.NoError(t, p.writeHandler(gossipapi.TxMetadata{BlockNum: 1, ChannelID: testChannel, TxID: "tx1"}, sideTreeTxnCCName, &kvrwset.KVWrite{Key: anchorAddrPrefix + k1, IsDelete: true, Value: []byte(v1)}))
		result := <-done
		require.Empty(t, result.AnchorAddress)
	})

	t.Run("test namespace not equal to sideTreeTxnCCName", func(t *testing.T) {
		// register to receive sidetree txn value
		sideTreeTxnCh := notifier.RegisterForSidetreeTxn()
		done := make(chan SideTreeTxn, 1)
		go func() {
			for {
				select {
				case sideTreeTxn := <-sideTreeTxnCh:
					done <- sideTreeTxn
				case <-time.After(1 * time.Second):
					done <- SideTreeTxn{}
				}
			}
		}()
		require.NoError(t, p.writeHandler(gossipapi.TxMetadata{BlockNum: 1, ChannelID: testChannel, TxID: "tx1"}, "n1", &kvrwset.KVWrite{Key: anchorAddrPrefix + k1, IsDelete: true, Value: []byte(v1)}))
		result := <-done
		require.Empty(t, result.AnchorAddress)
	})

	t.Run("test success", func(t *testing.T) {
		// register to receive sidetree txn value
		sideTreeTxnCh := notifier.RegisterForSidetreeTxn()
		done := make(chan SideTreeTxn, 1)
		go func() {
			for {
				select {
				case sideTreeTxn := <-sideTreeTxnCh:
					done <- sideTreeTxn
				case <-time.After(1 * time.Second):
					done <- SideTreeTxn{}
				}
			}
		}()
		require.NoError(t, p.writeHandler(gossipapi.TxMetadata{BlockNum: 1, ChannelID: testChannel, TxID: "tx1"}, sideTreeTxnCCName, &kvrwset.KVWrite{Key: anchorAddrPrefix + k1, IsDelete: false, Value: []byte(v1)}))
		result := <-done
		require.Equal(t, result.AnchorAddress, v1)
	})

}

type mockBlockPublisher struct {
	writeHandler gossipapi.WriteHandler
}

func (m *mockBlockPublisher) AddWriteHandler(writeHandler gossipapi.WriteHandler) {
	m.writeHandler = writeHandler

}
