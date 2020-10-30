/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package txn

import (
	"encoding/json"
	"fmt"
	"runtime/debug"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/common/flogging"
	ccapi "github.com/hyperledger/fabric/extensions/chaincode/api"
	dcasclient "github.com/trustbloc/fabric-peer-ext/pkg/collections/offledger/dcas/client"

	"github.com/trustbloc/sidetree-fabric/cmd/chaincode/cas"
	ctxcommon "github.com/trustbloc/sidetree-fabric/pkg/context/common"
	"github.com/trustbloc/sidetree-fabric/pkg/observer/common"
)

var logger = flogging.MustGetLogger("sidetreetxncc")

const (
	ccVersion = "v1"

	// Available function names
	writeContent = "writeContent"
	readContent  = "readContent"
	writeAnchor  = "writeAnchor"
	warmup       = "warmup"
)

// funcMap is a map of functions by function name
type funcMap map[string]func(shim.ChaincodeStubInterface, [][]byte) pb.Response

// ProtocolClientChannelProvider returns the protocol client provider for a given channel
type ProtocolClientChannelProvider interface {
	ProtocolClientProviderForChannel(channelID string) (ctxcommon.ProtocolClientProvider, error)
}

// DCASStubWrapperFactory creates a DCAS client wrapper around a stub
type DCASStubWrapperFactory interface {
	CreateDCASClientStubWrapper(coll string, stub shim.ChaincodeStubInterface) (dcasclient.DCAS, error)
}

// SidetreeTxnCC ...
type SidetreeTxnCC struct {
	name      string
	functions funcMap
	ProtocolClientChannelProvider
	DCASStubWrapperFactory
}

// New returns chaincode
func New(name string, pccp ProtocolClientChannelProvider, dcasClientFactory DCASStubWrapperFactory) *SidetreeTxnCC {
	cc := &SidetreeTxnCC{
		name:                          name,
		functions:                     make(funcMap),
		ProtocolClientChannelProvider: pccp,
		DCASStubWrapperFactory:        dcasClientFactory,
	}

	cc.functions[writeContent] = cc.write
	cc.functions[readContent] = cc.read
	cc.functions[writeAnchor] = cc.writeAnchor
	cc.functions[warmup] = cc.warmup

	return cc
}

// Name returns the name of this chaincode
func (cc *SidetreeTxnCC) Name() string { return cc.name }

// Version returns the version of this chaincode
func (cc *SidetreeTxnCC) Version() string { return ccVersion }

// Chaincode returns the SidetreeTxn chaincode
func (cc *SidetreeTxnCC) Chaincode() shim.Chaincode { return cc }

// GetDBArtifacts returns Couch DB indexes for the 'docs' collection
func (cc *SidetreeTxnCC) GetDBArtifacts([]string) map[string]*ccapi.DBArtifacts { return nil }

// Init - nothing to do for now
func (cc *SidetreeTxnCC) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

// Invoke manages metadata lifecycle operations (writeContent, readContent)
func (cc *SidetreeTxnCC) Invoke(stub shim.ChaincodeStubInterface) (resp pb.Response) {

	txID := stub.GetTxID()

	defer handlePanic(&resp)

	args := stub.GetArgs()
	if len(args) > 0 {
		// only display first arg (function), remaining args may contain client data, do not log them
		logger.Debugf("[txID %s] SidetreeTxnCC Arg[0]=%s", txID, args[0])
	}

	// Get function name (first argument)
	if len(args) < 1 {
		errMsg := "function name is required"
		logger.Debugf("[txID %s] %s", txID, errMsg)
		return shim.Error(errMsg)
	}

	functionName := string(args[0])
	function, valid := cc.functions[functionName]
	if !valid {
		errMsg := fmt.Sprintf("Invalid invoke function [%s]. Expecting one of: %s", functionName, cc.functions.String())
		logger.Debugf("[txID %s] %s", errMsg)
		return shim.Error(errMsg)
	}
	return function(stub, args[1:])
}

// writeContent will write content using cas client
func (cc *SidetreeTxnCC) write(stub shim.ChaincodeStubInterface, args [][]byte) pb.Response {
	txID := stub.GetTxID()
	if len(args) != 2 || len(args[0]) == 0 || len(args[1]) == 0 {
		err := "collection and content are required"
		logger.Debugf("[txID %s] %s", txID, err)
		return shim.Error(err)
	}

	dcasClient, err := cc.CreateDCASClientStubWrapper(string(args[0]), stub)
	if err != nil {
		errMsg := fmt.Sprintf("failed to create DCAS client: %s", err.Error())
		logger.Errorf("[txID %s] %s", txID, errMsg)
		return shim.Error(errMsg)
	}

	client := cas.New(dcasClient)

	address, err := client.Write(args[1])
	if err != nil {
		errMsg := fmt.Sprintf("failed to write content: %s", err.Error())
		logger.Errorf("[txID %s] %s", txID, errMsg)
		return shim.Error(errMsg)
	}

	return shim.Success([]byte(address))
}

// readContent will read content using cas client
func (cc *SidetreeTxnCC) read(stub shim.ChaincodeStubInterface, args [][]byte) pb.Response {
	txID := stub.GetTxID()
	if len(args) != 2 || len(args[0]) == 0 || len(args[1]) == 0 {
		errMsg := "collection and content address are required"
		logger.Debugf("[txID %s] %s", txID, errMsg)
		return shim.Error(errMsg)
	}

	dcasClient, err := cc.CreateDCASClientStubWrapper(string(args[0]), stub)
	if err != nil {
		errMsg := fmt.Sprintf("failed to create DCAS client: %s", err.Error())
		logger.Errorf("[txID %s] %s", txID, errMsg)
		return shim.Error(errMsg)
	}

	client := cas.New(dcasClient)

	address := string(args[1])
	payload, err := client.Read(address)
	if err != nil {
		errMsg := fmt.Sprintf("failed to read content: %s", err.Error())
		logger.Errorf("[txID %s] %s", txID, errMsg)
		return shim.Error(errMsg)
	}

	if payload == nil {
		return pb.Response{
			Status:  404,
			Message: "content not found",
		}
	}

	return shim.Success(payload)
}

// writeAnchor will record anchor info on the ledger
func (cc *SidetreeTxnCC) writeAnchor(stub shim.ChaincodeStubInterface, args [][]byte) pb.Response {
	txID := stub.GetTxID()
	if len(args) != 2 || len(args[0]) == 0 || len(args[1]) == 0 {
		errMsg := "missing anchor string and/or txn info"
		logger.Debugf("[txID %s] %s", txID, errMsg)
		return shim.Error(errMsg)
	}

	anchorString := string(args[0])
	txnInfoBytes := args[1]

	txnInfo := &common.TxnInfo{}
	err := json.Unmarshal(txnInfoBytes, txnInfo)
	if err != nil {
		errMsg := "invalid transaction info payload"
		logger.Debugf("[txID %s] %s: %s", txID, errMsg, err.Error())
		return shim.Error(errMsg)
	}

	err = cc.validateProtocolVersion(stub.GetChannelID(), txnInfo)
	if err != nil {
		return shim.Error(err.Error())
	}

	// record anchor string on the ledger plus Sidetree transaction info (anchor string, namespace)
	err = stub.PutState(common.AnchorPrefix+anchorString, txnInfoBytes)
	if err != nil {
		errMsg := fmt.Sprintf("failed to write anchor string: %s", err.Error())
		logger.Errorf("[txID %s] %s", txID, errMsg)
		return shim.Error(errMsg)
	}

	return shim.Success(nil)
}

func (cc *SidetreeTxnCC) validateProtocolVersion(channelID string, txnInfo *common.TxnInfo) error {
	pcp, err := cc.ProtocolClientProviderForChannel(channelID)
	if err != nil {
		return err
	}

	pc, err := pcp.ForNamespace(txnInfo.Namespace)
	if err != nil {
		return err
	}

	pv, err := pc.Get(txnInfo.ProtocolGenesisTime)
	if err != nil {
		return err
	}

	if txnInfo.ProtocolGenesisTime != pv.Protocol().GenesisTime {
		logger.Debugf("[%s] Request protocol genesis time [%d] ", channelID)

		return fmt.Errorf("invalid protocol genesis time in request: %d", txnInfo.ProtocolGenesisTime)
	}

	return nil
}

func (m funcMap) String() string {
	str := ""
	i := 0
	for key := range m {
		if i > 0 {
			str += ", "
		}
		i++
		str += fmt.Sprintf("\"%s\"", key)
	}
	return str
}

//nolint -- unused stub variable
func (cc *SidetreeTxnCC) warmup(stub shim.ChaincodeStubInterface, args [][]byte) pb.Response {
	return shim.Success(nil)
}

// handlePanic handles a panic (if any) by populating error response
func handlePanic(resp *pb.Response) {
	if r := recover(); r != nil {

		logger.Errorf("Recovering from panic: %s", string(debug.Stack()))

		errResp := shim.Error("panic: check server logs")
		resp.Reset()
		resp.Status = errResp.Status
		resp.Message = errResp.Message
	}
}
