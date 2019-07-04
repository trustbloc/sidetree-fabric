/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package observer

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/trustbloc/sidetree-core-go/pkg/api/batch"

	gossipapi "github.com/hyperledger/fabric/extensions/gossip/api"
	"github.com/hyperledger/fabric/protos/ledger/rwset/kvrwset"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	offledgerdcas "github.com/trustbloc/fabric-peer-ext/pkg/collections/offledger/dcas"
	"github.com/trustbloc/fabric-peer-ext/pkg/roles"
	sidetreeobserver "github.com/trustbloc/sidetree-core-go/pkg/observer"
	"github.com/trustbloc/sidetree-fabric/pkg/observer/config"
	dcasmocks "github.com/trustbloc/sidetree-fabric/pkg/observer/mocks"
)

const (
	confRoles = "ledger.roles"

	channel      = "diddoc"
	uniqueSuffix = "abc123"

	sideTreeTxnCCName = "sidetreetxn_cc"
	anchorAddrPrefix  = "sidetreetxn_"
	k1                = "key1"
)

func TestObserver(t *testing.T) {

	testRole := "endorser,observer"
	viper.Set(confRoles, testRole)

	client := getDefaultDCASClient()
	getDCAS = func(channelID string) dcasClient {
		return client
	}

	p := &mockBlockPublisher{}
	getBlockPublisher = func(channelID string) publisher {
		return p
	}

	cfg := config.New([]string{channel})
	require.Nil(t, Start(cfg))

	anchor := getAnchorAddress(uniqueSuffix)
	require.NoError(t, p.writeHandler(gossipapi.TxMetadata{BlockNum: 1, ChannelID: channel, TxID: "tx1"}, sideTreeTxnCCName, &kvrwset.KVWrite{Key: anchorAddrPrefix + k1, IsDelete: false, Value: []byte(anchor)}))
	time.Sleep(200 * time.Millisecond)

	// since there was one batch file with two operations we will have two entries in document map
	m, err := client.GetMap(docNs, docColl)
	require.Nil(t, err)
	require.Equal(t, len(m), 2)

}

func TestDCASPut(t *testing.T) {
	client := getDefaultDCASClient()
	client.PutErr = fmt.Errorf("put error")
	getDCAS = func(channelID string) dcasClient {
		return client
	}
	err := dcas{}.Put([]batch.Operation{{Type: "1"}})
	require.Error(t, err)
	require.Contains(t, err.Error(), "dcas put failed")
}

func TestObserverWithoutObserverRole(t *testing.T) {

	// create endorser role only
	rolesValue := make(map[roles.Role]struct{})
	rolesValue[roles.EndorserRole] = struct{}{}
	roles.SetRoles(rolesValue)

	cfg := config.New([]string{channel})
	require.Nil(t, Start(cfg))
}

func getDefaultDCASClient() *dcasmocks.MockDCASClient {

	client := dcasmocks.NewMockDCASClient()

	batchBytes, anchorBytes := getSidetreeTxnPrerequisites(uniqueSuffix)
	_, err := client.Put(sidetreeNs, sidetreeColl, batchBytes)
	if err != nil {
		panic(err)
	}

	_, err = client.Put(sidetreeNs, sidetreeColl, anchorBytes)
	if err != nil {
		panic(err)
	}

	return client
}

func getSidetreeTxnPrerequisites(uniqueSuffix string) (batchBytes, anchorBytes []byte) {

	operations := getDefaultOperations(uniqueSuffix)
	batchBytes = getBatchFileBytes(operations)
	batchAddr := offledgerdcas.GetCASKey(batchBytes)

	anchorBytes = getAnchorFileBytes(batchAddr, "")
	return batchBytes, anchorBytes
}

// Operation defines sample operation
type Operation struct {
	//Operation type
	Type string
	//The unique suffix - encoded hash of the original create document
	UniqueSuffix string
}

func getDefaultOperations(did string) []string {
	return []string{encode(Operation{UniqueSuffix: uniqueSuffix, Type: "create"}), encode(Operation{UniqueSuffix: uniqueSuffix, Type: "update"})}
}

func encode(op Operation) string {
	return base64.URLEncoding.EncodeToString([]byte(getJSON(op)))
}

func getJSON(op Operation) string {

	bytes, err := json.Marshal(op)
	if err != nil {
		panic(err)
	}

	return string(bytes)
}

func getBatchFileBytes(operations []string) []byte {
	bf := sidetreeobserver.BatchFile{Operations: operations}
	bytes, err := json.Marshal(bf)
	if err != nil {
		panic(err)
	}

	return bytes
}

func getAnchorFileBytes(batchFileHash string, merkleRoot string) []byte {
	af := sidetreeobserver.AnchorFile{
		BatchFileHash: batchFileHash,
		MerkleRoot:    merkleRoot,
	}
	s, err := json.Marshal(af)
	if err != nil {
		panic(err)
	}
	return s
}

type mockBlockPublisher struct {
	writeHandler gossipapi.WriteHandler
}

func (m *mockBlockPublisher) AddWriteHandler(writeHandler gossipapi.WriteHandler) {
	m.writeHandler = writeHandler
}

func getAnchorAddress(uniqueSuffix string) string {
	_, anchorBytes := getSidetreeTxnPrerequisites(uniqueSuffix)
	return offledgerdcas.GetCASKey(anchorBytes)
}
