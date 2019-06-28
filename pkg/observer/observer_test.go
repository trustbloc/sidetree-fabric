/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package observer

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/trustbloc/sidetree-core-go/pkg/docutil"

	gossipapi "github.com/hyperledger/fabric/extensions/gossip/api"
	"github.com/hyperledger/fabric/protos/ledger/rwset/kvrwset"

	"github.com/spf13/viper"
	"github.com/trustbloc/fabric-peer-ext/pkg/roles"

	"github.com/stretchr/testify/require"
	"github.com/trustbloc/fabric-peer-ext/pkg/collections/offledger/dcas"
	"github.com/trustbloc/sidetree-fabric/pkg/observer/config"
	"github.com/trustbloc/sidetree-fabric/pkg/observer/notifier"

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
	time.Sleep(1 * time.Second)

	// since there was one batch file with two operations we will have two entries in document map
	m, err := client.GetMap(docNs, docColl)
	require.Nil(t, err)
	require.Equal(t, len(m), 2)

}

func TestObserver_ProcessAnchorError(t *testing.T) {

	testRole := "endorser,observer"
	viper.Set(confRoles, testRole)

	p := &mockBlockPublisher{}
	getBlockPublisher = func(channelID string) publisher {
		return p
	}

	client := getDefaultDCASClient()
	getDCAS = func(channelID string) dcasClient {
		return client
	}

	client.GetErr = errors.New("get error")

	cfg := config.New([]string{channel})
	require.Nil(t, Start(cfg))

	require.NoError(t, p.writeHandler(gossipapi.TxMetadata{BlockNum: 1, ChannelID: channel, TxID: "tx1"}, sideTreeTxnCCName, &kvrwset.KVWrite{Key: anchorAddrPrefix + k1, IsDelete: false, Value: []byte("invalid")}))
	time.Sleep(1 * time.Second)

	// since there was an error during processing collection store will not be created and operation will not be added
	_, err := client.GetMap(docNs, docColl)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "collection store doesn't exist")
}

func TestObserver_DCASGetError(t *testing.T) {

	client := getDefaultDCASClient()
	client.GetErr = errors.New("get error")

	o := &observer{channelID: channel, dcas: client}
	err := o.processSidetreeTxn(notifier.SidetreeTxn{BlockNumber: 1, AnchorAddress: "n5BneBDAZMeJPMp9Kw0VTGGPckgIi5MGwSL_8VBE3qA=", TxNum: 1})
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "failed to retrieve content for anchor")

}

func TestObserver_DCASPutError(t *testing.T) {

	client := getDefaultDCASClient()
	client.PutErr = errors.New("put error")

	o := &observer{channelID: channel, dcas: client}
	err := o.processSidetreeTxn(getDefaultSidetreeTxn())
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "failed to store operation")

}

func TestGetAnchorFileError(t *testing.T) {

	getAnchorFile = func(bytes []byte) (*AnchorFile, error) {
		return nil, errors.New("anchor file error")
	}

	getBatchFile = func(bytes []byte) (*BatchFile, error) {
		return unmarshalBatchFile(bytes)
	}

	client := getDefaultDCASClient()
	o := &observer{channelID: channel, dcas: client}
	err := o.processSidetreeTxn(getDefaultSidetreeTxn())
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "anchor file error")

	// since there was an error during processing collection store will not be created and operation will not be added
	_, err = client.GetMap(docNs, docColl)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "collection store doesn't exist")

}

func TestGetBatchFileError(t *testing.T) {

	getBatchFile = func(bytes []byte) (*BatchFile, error) {
		return nil, errors.New("batch file error")
	}

	// original get anchor file function
	getAnchorFile = func(bytes []byte) (*AnchorFile, error) {
		return unmarshalAnchorFile(bytes)
	}

	client := getDefaultDCASClient()
	o := &observer{channelID: channel, dcas: client}
	err := o.processSidetreeTxn(getDefaultSidetreeTxn())
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "batch file error")

	// since there was one batch file with two operations we will have two entries in document map
	_, err = client.GetMap(docNs, docColl)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "collection store doesn't exist")

}

func TestObserverWithoutObserverRole(t *testing.T) {

	// create endorser role only
	rolesValue := make(map[roles.Role]struct{})
	rolesValue[roles.EndorserRole] = struct{}{}
	roles.SetRoles(rolesValue)

	cfg := config.New([]string{channel})
	require.Nil(t, Start(cfg))
}

func TestUnmarshalAnchorFile(t *testing.T) {
	af, err := unmarshalAnchorFile([]byte("[test : 123]"))
	require.NotNil(t, err)
	require.Nil(t, af)
	require.Contains(t, err.Error(), "invalid character")
}

func TestUnmarshalBatchFile(t *testing.T) {
	af, err := unmarshalBatchFile([]byte("[test : 123]"))
	require.NotNil(t, err)
	require.Nil(t, af)
	require.Contains(t, err.Error(), "invalid character")
}

func TestUpdateOperation(t *testing.T) {

	doc, err := updateOperation(docutil.EncodeToString([]byte("[test : 123]")), 1, notifier.SidetreeTxn{})
	require.NotNil(t, err)
	require.Nil(t, doc)
	require.Contains(t, err.Error(), "invalid character")
}

func getJSON(op Operation) string {

	bytes, err := json.Marshal(op)
	if err != nil {
		panic(err)
	}

	return string(bytes)
}

// Operation defines sample operation
type Operation struct {
	//Operation type
	Type string
	//The unique suffix - encoded hash of the original create document
	UniqueSuffix string
}

func encode(op Operation) string {
	return base64.URLEncoding.EncodeToString([]byte(getJSON(op)))
}

func getDefaultOperations(did string) []string {
	return []string{encode(Operation{UniqueSuffix: uniqueSuffix, Type: "create"}), encode(Operation{UniqueSuffix: uniqueSuffix, Type: "update"})}
}

func getBatchFileBytes(operations []string) []byte {
	bf := BatchFile{Operations: operations}
	bytes, err := json.Marshal(bf)
	if err != nil {
		panic(err)
	}

	return bytes
}

func getSidetreeTxnPrerequisites(uniqueSuffix string) (batchBytes, anchorBytes []byte) {

	operations := getDefaultOperations(uniqueSuffix)
	batchBytes = getBatchFileBytes(operations)
	batchAddr := dcas.GetCASKey(batchBytes)

	anchorBytes = getAnchorFileBytes(batchAddr, "")
	return batchBytes, anchorBytes
}

func getAnchorAddress(uniqueSuffix string) string {
	_, anchorBytes := getSidetreeTxnPrerequisites(uniqueSuffix)
	return dcas.GetCASKey(anchorBytes)
}

func getDefaultSidetreeTxn() notifier.SidetreeTxn {
	return notifier.SidetreeTxn{BlockNumber: 1, AnchorAddress: getAnchorAddress(uniqueSuffix), TxNum: 1}
}

func getAnchorFileBytes(batchFileHash string, merkleRoot string) []byte {
	af := AnchorFile{
		BatchFileHash: batchFileHash,
		MerkleRoot:    merkleRoot,
	}
	s, err := json.Marshal(af)
	if err != nil {
		panic(err)
	}
	return s
}

var getMockDCAS = func(channelID string) dcasClient {
	return getDefaultDCASClient()
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

type mockBlockPublisher struct {
	writeHandler gossipapi.WriteHandler
}

func (m *mockBlockPublisher) AddWriteHandler(writeHandler gossipapi.WriteHandler) {
	m.writeHandler = writeHandler
}
