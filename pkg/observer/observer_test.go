/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package observer

import (
	"encoding/base64"
	"encoding/json"
	"testing"
	"time"

	"github.com/hyperledger/fabric-protos-go/ledger/rwset/kvrwset"
	gossipapi "github.com/hyperledger/fabric/extensions/gossip/api"
	"github.com/stretchr/testify/require"
	offledgerdcas "github.com/trustbloc/fabric-peer-ext/pkg/collections/offledger/dcas"
	"github.com/trustbloc/fabric-peer-ext/pkg/mocks"
	extroles "github.com/trustbloc/fabric-peer-ext/pkg/roles"
	"github.com/trustbloc/sidetree-core-go/pkg/api/batch"
	"github.com/trustbloc/sidetree-core-go/pkg/docutil"
	sidetreeobserver "github.com/trustbloc/sidetree-core-go/pkg/observer"
	"github.com/trustbloc/sidetree-core-go/pkg/patch"
	"github.com/trustbloc/sidetree-core-go/pkg/restapi/model"

	"github.com/trustbloc/sidetree-fabric/pkg/config"
	stmocks "github.com/trustbloc/sidetree-fabric/pkg/mocks"
	"github.com/trustbloc/sidetree-fabric/pkg/observer/common"
	obmocks "github.com/trustbloc/sidetree-fabric/pkg/observer/mocks"
	"github.com/trustbloc/sidetree-fabric/pkg/observer/notifier"
	"github.com/trustbloc/sidetree-fabric/pkg/role"
)

//go:generate counterfeiter -o ../mocks/operationstore.gen.go --fake-name OperationStore ../context/common OperationStore
//go:generate counterfeiter -o ../mocks/opstoreprovider.gen.go --fake-name OperationStoreProvider ../context/common OperationStoreProvider

const (
	channel      = "diddoc"
	namespace    = "did:sidetree"
	uniqueSuffix = "abc123"

	sideTreeTxnCCName = "sidetreetxn_cc"
	dcasColl          = "dcas"
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
	bpp := mocks.NewBlockPublisherProvider().WithBlockPublisher(p)

	dcasCfg := config.DCAS{
		ChaincodeName: sideTreeTxnCCName,
		Collection:    dcasColl,
	}

	c := getDefaultDCASClient(dcasCfg)
	dcasProvider := &stmocks.DCASClientProvider{}
	dcasProvider.ForChannelReturns(c, nil)

	opStore := &stmocks.OperationStore{}
	opStoreProvider := &stmocks.OperationStoreProvider{}
	opStoreProvider.ForNamespaceReturns(opStore, nil)

	observer := New(channel, dcasCfg,
		&Providers{
			DCAS:           dcasProvider,
			OperationStore: opStoreProvider,
			Ledger:         notifier.New(channel, bpp),
			Filter:         &sidetreeobserver.NoopOperationFilterProvider{},
		},
	)
	require.NotNil(t, observer)
	require.NoError(t, observer.Start())

	defer observer.Stop()

	txMetaData := gossipapi.TxMetadata{BlockNum: 1, ChannelID: channel, TxID: "tx1"}
	kvWrite := &kvrwset.KVWrite{Key: common.AnchorAddrPrefix + k1, IsDelete: false, Value: []byte(getAnchorAddress(uniqueSuffix))}

	require.NoError(t, p.HandleWrite(txMetaData, sideTreeTxnCCName, kvWrite))
	time.Sleep(200 * time.Millisecond)

	// since there was one batch file with two operations we will have two entries in document map
	putCalls := opStore.Invocations()["Put"]
	require.Len(t, putCalls, 1)
	require.Len(t, putCalls[0], 1)

	ops, ok := putCalls[0][0].([]*batch.Operation)
	require.True(t, ok)
	require.Len(t, ops, 2)
}

func getDefaultDCASClient(cfg config.DCAS) *obmocks.MockDCASClient {
	dcasClient := obmocks.NewMockDCASClient()

	batchBytes, anchorBytes := getSidetreeTxnPrerequisites(uniqueSuffix)
	_, err := dcasClient.Put(cfg.ChaincodeName, cfg.Collection, batchBytes)
	if err != nil {
		panic(err)
	}

	_, err = dcasClient.Put(cfg.ChaincodeName, cfg.Collection, anchorBytes)
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
	//Operation patch data
	PatchData *model.PatchDataModel
}

func getDefaultOperations(did string) []string {
	id := namespace + docutil.NamespaceDelimiter + uniqueSuffix

	return []string{
		encode(Operation{ID: id, UniqueSuffix: uniqueSuffix, Type: "create", PatchData: &model.PatchDataModel{
			Patches: []patch.Patch{},
		}}),
		encode(Operation{ID: id, UniqueSuffix: uniqueSuffix, Type: "update", PatchData: &model.PatchDataModel{
			Patches: []patch.Patch{},
		}}),
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
