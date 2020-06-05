/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package notifier

import (
	"testing"
	"time"

	"github.com/hyperledger/fabric-protos-go/ledger/rwset/kvrwset"
	gossipapi "github.com/hyperledger/fabric/extensions/gossip/api"
	"github.com/stretchr/testify/require"
	"github.com/trustbloc/fabric-peer-ext/pkg/mocks"

	"github.com/trustbloc/sidetree-fabric/pkg/observer/common"
)

const (
	testChannel       = "testChannel"
	k1                = "key1"
	v1                = "value1"
	sideTreeTxnCCName = "sidetreetxn_cc"
)

func TestRegisterForAnchorFileAddress(t *testing.T) {
	p := mocks.NewBlockPublisher()
	provider := mocks.NewBlockPublisherProvider().WithBlockPublisher(p)

	sideTreeTxnCh := make(chan gossipapi.TxMetadata, 1)

	t.Run("test key in kvrwset is deleted", func(t *testing.T) {
		notifier := New(testChannel, provider, sideTreeTxnCh)
		require.NotNil(t, notifier)

		done := make(chan gossipapi.TxMetadata, 1)
		go func() {
			for {
				select {
				case sideTreeTxn := <-sideTreeTxnCh:
					done <- sideTreeTxn
				case <-time.After(1 * time.Second):
					done <- gossipapi.TxMetadata{}
					return
				}
			}
		}()
		require.NoError(t, p.HandleWrite(gossipapi.TxMetadata{BlockNum: 1, ChannelID: testChannel, TxID: "tx1"}, sideTreeTxnCCName, &kvrwset.KVWrite{Key: common.AnchorPrefix + k1, IsDelete: true, Value: []byte(v1)}))
		result := <-done
		require.Empty(t, result)
	})

	t.Run("test success", func(t *testing.T) {
		notifier := New(testChannel, provider, sideTreeTxnCh)
		require.NotNil(t, notifier)

		done := make(chan gossipapi.TxMetadata, 1)
		go func() {
			for {
				select {
				case sideTreeTxn := <-sideTreeTxnCh:
					done <- sideTreeTxn
				case <-time.After(1 * time.Second):
					done <- gossipapi.TxMetadata{}
					return
				}
			}
		}()
		require.NoError(t, p.HandleWrite(gossipapi.TxMetadata{BlockNum: 1, ChannelID: testChannel, TxID: "tx1"}, sideTreeTxnCCName, &kvrwset.KVWrite{Key: common.AnchorPrefix + k1, IsDelete: false, Value: []byte(v1)}))
		result := <-done
		require.Equal(t, uint64(1), result.BlockNum)
	})
}
