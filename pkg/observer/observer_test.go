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

	"github.com/hyperledger/fabric-protos-go/ledger/rwset/kvrwset"
	gossipapi "github.com/hyperledger/fabric/extensions/gossip/api"
	viper "github.com/spf13/viper2015"
	"github.com/stretchr/testify/require"
	offledgerdcas "github.com/trustbloc/fabric-peer-ext/pkg/collections/offledger/dcas"
	dcasclient "github.com/trustbloc/fabric-peer-ext/pkg/collections/offledger/dcas/client"
	"github.com/trustbloc/fabric-peer-ext/pkg/mocks"
	extroles "github.com/trustbloc/fabric-peer-ext/pkg/roles"
	"github.com/trustbloc/sidetree-core-go/pkg/api/batch"
	"github.com/trustbloc/sidetree-core-go/pkg/docutil"
	sidetreeobserver "github.com/trustbloc/sidetree-core-go/pkg/observer"
	stmocks "github.com/trustbloc/sidetree-fabric/pkg/mocks"
	"github.com/trustbloc/sidetree-fabric/pkg/observer/common"
	obmocks "github.com/trustbloc/sidetree-fabric/pkg/observer/mocks"
	"github.com/trustbloc/sidetree-fabric/pkg/role"
)

const (
	channel      = "diddoc"
	namespace    = "did:sidetree"
	uniqueSuffix = "abc123"

	sideTreeTxnCCName = "sidetreetxn_cc"
	anchorAddrPrefix  = "sidetreetxn_"
	k1                = "key1"
)

func TestObserver(t *testing.T) {
	rolesValue := make(map[extroles.Role]struct{})
	rolesValue[extroles.EndorserRole] = struct{}{}
	rolesValue[role.Observer] = struct{}{}
	extroles.SetRoles(rolesValue)
	defer func() {
		extroles.SetRoles(nil)
	}()

	p := mocks.NewBlockPublisher()

	c := getDefaultDCASClient()
	dcasProvider := &stmocks.DCASClientProvider{}
	dcasProvider.ForChannelReturns(c, nil)

	opStore := &obmocks.OperationStoreClient{}
	opStoreProvider := &obmocks.OpStoreClientProvider{}
	opStoreProvider.GetReturns(opStore)

	providers := &Providers{
		DCAS:            dcasProvider,
		OffLedger:       &obmocks.OffLedgerClientProvider{},
		BlockPublisher:  mocks.NewBlockPublisherProvider().WithBlockPublisher(p),
		OpStoreProvider: opStoreProvider,
	}
	observer := New(channel, providers)
	require.NotNil(t, observer)
	require.NoError(t, observer.Start())

	anchor := getAnchorAddress(uniqueSuffix)
	require.NoError(t, p.HandleWrite(gossipapi.TxMetadata{BlockNum: 1, ChannelID: channel, TxID: "tx1"}, sideTreeTxnCCName, &kvrwset.KVWrite{Key: anchorAddrPrefix + k1, IsDelete: false, Value: []byte(anchor)}))
	time.Sleep(200 * time.Millisecond)

	// since there was one batch file with two operations we will have two entries in document map
	m, err := c.GetMap(common.DocNs, common.DocColl)
	require.Nil(t, err)
	require.Equal(t, 2, len(m))
}

func TestDCASPut(t *testing.T) {
	c := getDefaultDCASClient()
	c.PutErr = fmt.Errorf("put error")
	dcasClientProvider := &mockDCASClientProvider{
		client: c,
	}
	err := (newDCAS(channel, dcasClientProvider)).Put([]*batch.Operation{{Type: "1"}})
	require.Error(t, err)
	require.Contains(t, err.Error(), "dcas put failed")
}

func TestObserver_Start(t *testing.T) {
	// Ensure the roles are initialized, otherwise they'll be overwritten
	// when we run the tests
	require.True(t, extroles.IsEndorser())

	t.Run("endorser role", func(t *testing.T) {
		// create endorser role only
		rolesValue := make(map[extroles.Role]struct{})
		rolesValue[extroles.EndorserRole] = struct{}{}
		extroles.SetRoles(rolesValue)
		defer func() {
			extroles.SetRoles(nil)
		}()

		providers := &Providers{
			DCAS:           &stmocks.DCASClientProvider{},
			OffLedger:      &obmocks.OffLedgerClientProvider{},
			BlockPublisher: mocks.NewBlockPublisherProvider(),
		}
		observer := New(channel, providers)
		require.NotNil(t, observer)

		observer.Start()

		observer.Stop()
	})

	t.Run("monitor role", func(t *testing.T) {
		rolesValue := make(map[extroles.Role]struct{})
		rolesValue[extroles.CommitterRole] = struct{}{}
		rolesValue[role.Resolver] = struct{}{}
		extroles.SetRoles(rolesValue)
		defer func() {
			extroles.SetRoles(nil)
		}()

		providers := &Providers{
			DCAS:           &stmocks.DCASClientProvider{},
			OffLedger:      &obmocks.OffLedgerClientProvider{},
			BlockPublisher: mocks.NewBlockPublisherProvider(),
		}
		observer := New(channel, providers)
		require.NotNil(t, observer)

		viper.Set("peer.id", "peer0.org1.com")
		observer.Start()
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

	anchorBytes = getAnchorFileBytes(batchAddr, []string{uniqueSuffix})
	return batchBytes, anchorBytes
}

// Operation defines sample operation
type Operation struct {
	//Operation type
	Type string
	//The unique suffix - encoded hash of the original create document
	UniqueSuffix string
	//The full ID of the document (including namespace)
	ID string
}

func getDefaultOperations(did string) []string {
	id := namespace + docutil.NamespaceDelimiter + uniqueSuffix

	return []string{
		encode(Operation{ID: id, UniqueSuffix: uniqueSuffix, Type: "create"}),
		encode(Operation{ID: id, UniqueSuffix: uniqueSuffix, Type: "update"}),
	}
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

func getAnchorFileBytes(batchFileHash string, uniqueSuffixes []string) []byte {
	af := sidetreeobserver.AnchorFile{
		BatchFileHash:  batchFileHash,
		UniqueSuffixes: uniqueSuffixes,
	}
	_, bytes, err := common.MarshalDCAS(af)
	if err != nil {
		panic(err)
	}
	return bytes
}

func getAnchorAddress(uniqueSuffix string) string {
	_, anchorBytes := getSidetreeTxnPrerequisites(uniqueSuffix)
	key, _, err := offledgerdcas.GetCASKeyAndValue(anchorBytes)
	if err != nil {
		panic(err.Error())
	}
	return key
}

func TestMain(t *testing.M) {
	// Ensure that the roles are pre-initialized
	extroles.GetRoles()
	t.Run()
}
