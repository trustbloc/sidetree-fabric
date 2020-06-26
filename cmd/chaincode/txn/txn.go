/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package txn

import (
	"fmt"
	"runtime/debug"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/common/flogging"
	ccapi "github.com/hyperledger/fabric/extensions/chaincode/api"

	"github.com/trustbloc/sidetree-fabric/cmd/chaincode/cas"
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

// SidetreeTxnCC ...
type SidetreeTxnCC struct {
	name      string
	functions funcMap
}

// New returns chaincode
func New(name string) *SidetreeTxnCC {
	cc := &SidetreeTxnCC{
		name:      name,
		functions: make(funcMap),
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

	client := cas.New(stub, string(args[0]))

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

	client := cas.New(stub, string(args[0]))

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
	txnInfo := args[1]

	// record anchor string on the ledger plus Sidetree transaction info (anchor string, namespace)
	err := stub.PutState(common.AnchorPrefix+anchorString, txnInfo)
	if err != nil {
		errMsg := fmt.Sprintf("failed to write anchor string: %s", err.Error())
		logger.Errorf("[txID %s] %s", txID, errMsg)
		return shim.Error(errMsg)
	}

	return shim.Success(nil)
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
