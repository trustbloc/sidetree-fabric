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

	"github.com/trustbloc/fabric-peer-ext/pkg/roles"
	"github.com/trustbloc/sidetree-core-go/pkg/api/batch"
	"github.com/trustbloc/sidetree-fabric/pkg/client"
	"github.com/trustbloc/sidetree-fabric/pkg/observer/common"

	gossipapi "github.com/hyperledger/fabric/extensions/gossip/api"
	"github.com/hyperledger/fabric/protos/ledger/rwset/kvrwset"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	offledgerdcas "github.com/trustbloc/fabric-peer-ext/pkg/collections/offledger/dcas"
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

	c := getDefaultDCASClient()
	getDCASClientProvider = func() dcasClientProvider {
		return &mockDCASClientProvider{
			client: c,
		}
	}

	p := &mockBlockPublisher{}
	getBlockPublisher = func(channelID string) publisher {
		return p
	}

	cfg := config.New([]string{channel})
	observer := New(cfg)
	require.NotNil(t, observer)
	require.NoError(t, observer.Start())

	anchor := getAnchorAddress(uniqueSuffix)
	require.NoError(t, p.writeHandler(gossipapi.TxMetadata{BlockNum: 1, ChannelID: channel, TxID: "tx1"}, sideTreeTxnCCName, &kvrwset.KVWrite{Key: anchorAddrPrefix + k1, IsDelete: false, Value: []byte(anchor)}))
	time.Sleep(200 * time.Millisecond)

	// since there was one batch file with two operations we will have two entries in document map
	m, err := c.GetMap(common.DocNs, common.DocColl)
	require.Nil(t, err)
	require.Equal(t, len(m), 2)

}

func TestDCASPut(t *testing.T) {
	c := getDefaultDCASClient()
	c.PutErr = fmt.Errorf("put error")
	dcasClientProvider := &mockDCASClientProvider{
		client: c,
	}
	err := (newDCAS(channel, dcasClientProvider)).Put([]batch.Operation{{Type: "1"}})
	require.Error(t, err)
	require.Contains(t, err.Error(), "dcas put failed")
}

func TestObserver_Start(t *testing.T) {
	// Ensure the roles are initialized, otherwise they'll be overwritten
	// when we run the tests
	require.True(t, roles.IsEndorser())

	t.Run("endorser role", func(t *testing.T) {
		// create endorser role only
		rolesValue := make(map[roles.Role]struct{})
		rolesValue[roles.EndorserRole] = struct{}{}
		roles.SetRoles(rolesValue)
		defer func() {
			roles.SetRoles(nil)
		}()

		cfg := config.New([]string{channel})
		observer := New(cfg)
		require.NotNil(t, observer)
		require.NoError(t, observer.Start())
	})
	t.Run("committer role", func(t *testing.T) {
		rolesValue := make(map[roles.Role]struct{})
		rolesValue[roles.CommitterRole] = struct{}{}
		roles.SetRoles(rolesValue)
		defer func() {
			roles.SetRoles(nil)
		}()

		cfg := config.New([]string{channel})
		observer := New(cfg)
		require.NotNil(t, observer)
		require.NoError(t, observer.Start())
	})
}

type mockDCASClientProvider struct {
	client *dcasmocks.MockDCASClient
}

func (m *mockDCASClientProvider) ForChannel(channelID string) client.DCAS {
	return m.client
}

func getDefaultDCASClient() *dcasmocks.MockDCASClient {

	dcasClient := dcasmocks.NewMockDCASClient()

	batchBytes, anchorBytes := getSidetreeTxnPrerequisites(uniqueSuffix)
	_, err := dcasClient.Put(common.SidetreeNs, common.SidetreeColl, batchBytes)
	if err != nil {
		panic(err)
	}

	_, err = dcasClient.Put(common.SidetreeNs, common.SidetreeColl, anchorBytes)
	if err != nil {
		panic(err)
	}

	return dcasClient
}

func getSidetreeTxnPrerequisites(uniqueSuffix string) (batchBytes, anchorBytes []byte) {

	operations := getDefaultOperations(uniqueSuffix)
	batchAddr, batchBytes := getBatchFileBytes(operations)

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

func getBatchFileBytes(operations []string) (string, []byte) {
	bf := sidetreeobserver.BatchFile{Operations: operations}
	key, bytes, err := common.MarshalDCAS(bf)
	if err != nil {
		panic(err)
	}
	return key, bytes
}

func getAnchorFileBytes(batchFileHash string, merkleRoot string) []byte {
	af := sidetreeobserver.AnchorFile{
		BatchFileHash: batchFileHash,
		MerkleRoot:    merkleRoot,
	}
	_, bytes, err := common.MarshalDCAS(af)
	if err != nil {
		panic(err)
	}
	return bytes
}

type mockBlockPublisher struct {
	writeHandler gossipapi.WriteHandler
}

func (m *mockBlockPublisher) AddWriteHandler(writeHandler gossipapi.WriteHandler) {
	m.writeHandler = writeHandler
}

func getAnchorAddress(uniqueSuffix string) string {
	_, anchorBytes := getSidetreeTxnPrerequisites(uniqueSuffix)
	key, _, err := offledgerdcas.GetCASKeyAndValue(anchorBytes)
	if err != nil {
		panic(err.Error())
	}
	return key
}
