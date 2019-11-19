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

	gossipapi "github.com/hyperledger/fabric/extensions/gossip/api"
	"github.com/hyperledger/fabric/protos/ledger/rwset/kvrwset"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	offledgerdcas "github.com/trustbloc/fabric-peer-ext/pkg/collections/offledger/dcas"
	dcasclient "github.com/trustbloc/fabric-peer-ext/pkg/collections/offledger/dcas/client"
	"github.com/trustbloc/fabric-peer-ext/pkg/mocks"
	"github.com/trustbloc/fabric-peer-ext/pkg/roles"
	"github.com/trustbloc/sidetree-core-go/pkg/api/batch"
	sidetreeobserver "github.com/trustbloc/sidetree-core-go/pkg/observer"
	"github.com/trustbloc/sidetree-fabric/pkg/observer/common"
	"github.com/trustbloc/sidetree-fabric/pkg/observer/config"
	obmocks "github.com/trustbloc/sidetree-fabric/pkg/observer/mocks"
)

const (
	confRoles = "ledger.roles"

	channel      = "diddoc"
	uniqueSuffix = "abc123"

	sideTreeTxnCCName = "sidetreetxn_cc"
	anchorAddrPrefix  = "sidetreetxn_"
	k1                = "key1"

	monitorPeriod = 5 * time.Second
)

func TestObserver(t *testing.T) {

	testRole := "endorser,observer"
	viper.Set(confRoles, testRole)

	p := mocks.NewBlockPublisher()

	c := getDefaultDCASClient()
	dcasProvider := &obmocks.DCASClientProvider{}
	dcasProvider.ForChannelReturns(c, nil)

	cfg := config.New([]string{channel}, monitorPeriod)

	providers := &Providers{
		DCAS:           dcasProvider,
		OffLedger:      &obmocks.OffLedgerClientProvider{},
		BlockPublisher: mocks.NewBlockPublisherProvider().WithBlockPublisher(p),
	}
	observer := New(cfg, providers)
	require.NotNil(t, observer)
	require.NoError(t, observer.Start())

	anchor := getAnchorAddress(uniqueSuffix)
	require.NoError(t, p.HandleWrite(gossipapi.TxMetadata{BlockNum: 1, ChannelID: channel, TxID: "tx1"}, sideTreeTxnCCName, &kvrwset.KVWrite{Key: anchorAddrPrefix + k1, IsDelete: false, Value: []byte(anchor)}))
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

		cfg := config.New([]string{channel}, monitorPeriod)
		providers := &Providers{
			DCAS:           &obmocks.DCASClientProvider{},
			OffLedger:      &obmocks.OffLedgerClientProvider{},
			BlockPublisher: mocks.NewBlockPublisherProvider(),
		}
		observer := New(cfg, providers)
		require.NotNil(t, observer)
		require.NoError(t, observer.Start())
		observer.Stop()
	})
	t.Run("monitor role", func(t *testing.T) {
		rolesValue := make(map[roles.Role]struct{})
		rolesValue[roles.CommitterRole] = struct{}{}
		rolesValue[sidetreeRole] = struct{}{}
		roles.SetRoles(rolesValue)
		defer func() {
			roles.SetRoles(nil)
		}()

		cfg := config.New([]string{channel}, monitorPeriod)
		providers := &Providers{
			DCAS:           &obmocks.DCASClientProvider{},
			OffLedger:      &obmocks.OffLedgerClientProvider{},
			BlockPublisher: mocks.NewBlockPublisherProvider(),
		}
		observer := New(cfg, providers)
		require.NotNil(t, observer)

		err := observer.Start()
		require.Error(t, err)
		require.Contains(t, err.Error(), "peer.id isn't set")

		viper.Set("peer.id", "peer0.org1.com")
		require.NoError(t, observer.Start())
		observer.Stop()
	})
}

type mockDCASClientProvider struct {
	client dcasclient.DCAS
}

func (m *mockDCASClientProvider) ForChannel(channelID string) (dcasclient.DCAS, error) {
	return m.client, nil
}

func getDefaultDCASClient() *obmocks.MockDCASClient {
	dcasClient := obmocks.NewMockDCASClient()

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
