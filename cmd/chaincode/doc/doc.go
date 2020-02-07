/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package doc

import (
	"encoding/json"
	"fmt"
	"runtime/debug"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	ccapi "github.com/hyperledger/fabric/extensions/chaincode/api"
	"github.com/trustbloc/sidetree-fabric/cmd/chaincode/cas"
)

var logger = NewLogger("doc")

const (
	ccVersion = "v1"

	// Available function names
	write     = "write"
	read      = "read"
	queryByID = "queryByID"
	warmup    = "warmup"

	// this chaincode can be 'generic' if we pass in collection name to each function
	collection = "docs"

	couchDB           = "couchdb"
	docsCollIndex     = `{"index": {"fields": ["id"]}, "ddoc": "indexIDDoc", "name": "indexID", "type": "json"}`
	queryByIDTemplate = `{"selector":{"id":"%s"},"use_index":["_design/indexIDDoc","indexID"],"fields":["id","encodedPayload","hashAlgorithmInMultiHashCode","operationIndex","operationNumber","patch","previousOperationHash","signature","signingKeyID","transactionNumber","transactionTime","type","uniqueSuffix"]}`
)

// funcMap is a map of functions by function name
type funcMap map[string]func(shim.ChaincodeStubInterface, [][]byte) pb.Response

// DocumentCC ...
type DocumentCC struct {
	name      string
	functions funcMap
}

// New returns chaincode
func New(name string) *DocumentCC {
	cc := &DocumentCC{
		name:      name,
		functions: make(funcMap),
	}

	cc.functions[write] = cc.write
	cc.functions[read] = cc.read
	cc.functions[queryByID] = cc.queryByID
	cc.functions[warmup] = cc.warmup

	return cc
}

// Name returns the name of this chaincode
func (cc *DocumentCC) Name() string { return cc.name }

// Version returns the version of this chaincode
func (cc *DocumentCC) Version() string { return ccVersion }

// Chaincode returns the DocumentCC chaincode
func (cc *DocumentCC) Chaincode() shim.Chaincode { return cc }

// GetDBArtifacts returns Couch DB indexes for the 'docs' collection
func (cc *DocumentCC) GetDBArtifacts() map[string]*ccapi.DBArtifacts {
	return map[string]*ccapi.DBArtifacts{
		couchDB: {
			CollectionIndexes: map[string][]string{
				collection: {docsCollIndex},
			},
		},
	}
}

// Init - nothing to do for now
func (cc *DocumentCC) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

// Invoke manages document lifecycle operations (write, read)
func (cc *DocumentCC) Invoke(stub shim.ChaincodeStubInterface) (resp pb.Response) {

	txID := stub.GetTxID()

	defer handlePanic(&resp)

	args := stub.GetArgs()
	if len(args) > 0 {
		// only display first arg (function), remaining args may contain client data, do not log them
		logger.Debugf("[txID %s] DocumentCC Arg[0]=%s", txID, args[0])
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
		errMsg := fmt.Sprintf("invalid invoke function [%s] - expecting one of: %s", functionName, cc.functions.String())
		logger.Debugf("[txID %s] %s", errMsg)
		return shim.Error(errMsg)
	}
	return function(stub, args[1:])
}

// write will write content using cas client
func (cc *DocumentCC) write(stub shim.ChaincodeStubInterface, args [][]byte) pb.Response {
	txID := stub.GetTxID()
	if len(args) < 1 || len(args[0]) == 0 {
		err := "missing content"
		logger.Debugf("[txID %s] %s", txID, err)
		return shim.Error(err)
	}

	casClient := cas.New(stub, collection)
	address, err := casClient.Write(args[0])
	if err != nil {
		errMsg := fmt.Sprintf("failed to write content: %s", err.Error())
		logger.Errorf("[txID %s] %s", txID, errMsg)
		return shim.Error(errMsg)
	}

	return shim.Success([]byte(address))
}

// read will read content using cas client
func (cc *DocumentCC) read(stub shim.ChaincodeStubInterface, args [][]byte) pb.Response {
	txID := stub.GetTxID()
	if len(args) < 1 || len(args[0]) == 0 {
		errMsg := "missing content address"
		logger.Debugf("[txID %s] %s", txID, errMsg)
		return shim.Error(errMsg)
	}

	client := cas.New(stub, collection)

	address := string(args[0])
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

// queryByID wiil query all operations for document with specified ID
func (cc *DocumentCC) queryByID(stub shim.ChaincodeStubInterface, args [][]byte) pb.Response {

	txID := stub.GetTxID()
	if len(args) < 1 {
		err := "id is required"
		logger.Debugf("[txID %s] %s", txID, err)
		return shim.Error(err)
	}

	ID := string(args[0])

	client := cas.New(stub, collection)
	operations, err := client.Query(fmt.Sprintf(queryByIDTemplate, ID))
	if err != nil {
		errMsg := fmt.Sprintf("failed to query operations by id[%s]: %s", ID, err.Error())
		logger.Errorf("[txID %s] %s", txID, errMsg)
		return shim.Error(errMsg)
	}

	if len(operations) == 0 {
		return pb.Response{
			Status:  404,
			Message: "document not found",
		}
	}

	// TODO: sort documents by blockchain time (block number, tx number within block, operation index within batch)

	payload, err := json.Marshal(operations)
	if err != nil {
		errMsg := fmt.Sprintf("failed to marshal documents: %s", err.Error())
		logger.Errorf("[txID %s] %s", txID, errMsg)
		return shim.Error(errMsg)
	}

	return shim.Success(payload)
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
func (cc *DocumentCC) warmup(stub shim.ChaincodeStubInterface, args [][]byte) pb.Response {
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
