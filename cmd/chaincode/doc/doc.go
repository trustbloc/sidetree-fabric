/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"encoding/json"
	"fmt"
	"runtime/debug"

	"github.com/trustbloc/sidetree-fabric/cmd/chaincode/cas"

	"github.com/hyperledger/fabric/core/chaincode/shim"

	pb "github.com/hyperledger/fabric/protos/peer"
)

var logger = shim.NewLogger("doc")

const (
	// Available function names
	write     = "write"
	read      = "read"
	queryByID = "queryByID"
	warmup    = "warmup"

	// this chaincode can be 'generic' if we pass in collection name to each function
	collection = "docs"

	queryByIDTemplate = `{"selector":{"id":"%s"},"use_index":["_design/indexIDDoc","indexID"],"fields":["id","encodedPayload","hashAlgorithmInMultiHashCode","operationIndex","operationNumber","patch","previousOperationHash","signature","signingKeyID","transactionNumber","transactionTime","type","uniqueSuffix"]}`
)

// funcMap is a map of functions by function name
type funcMap map[string]func(shim.ChaincodeStubInterface, [][]byte) pb.Response

// DocumentCC ...
type DocumentCC struct {
	functions funcMap
}

// new returns chaincode
func new() shim.Chaincode {

	cc := &DocumentCC{functions: make(funcMap)}

	cc.functions[write] = cc.write
	cc.functions[read] = cc.read
	cc.functions[queryByID] = cc.queryByID
	cc.functions[warmup] = cc.warmup

	return cc
}

// Init - nothing to do for now
func (t *DocumentCC) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

// Invoke manages document lifecycle operations (write, read)
func (t *DocumentCC) Invoke(stub shim.ChaincodeStubInterface) (resp pb.Response) {

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
	function, valid := t.functions[functionName]
	if !valid {
		errMsg := fmt.Sprintf("invalid invoke function [%s] - expecting one of: %s", functionName, t.functions.String())
		logger.Debugf("[txID %s] %s", errMsg)
		return shim.Error(errMsg)
	}
	return function(stub, args[1:])
}

// write will write content using cas client
func (t *DocumentCC) write(stub shim.ChaincodeStubInterface, args [][]byte) pb.Response {
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
func (t *DocumentCC) read(stub shim.ChaincodeStubInterface, args [][]byte) pb.Response {
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
func (t *DocumentCC) queryByID(stub shim.ChaincodeStubInterface, args [][]byte) pb.Response {

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
func (t *DocumentCC) warmup(stub shim.ChaincodeStubInterface, args [][]byte) pb.Response {
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

func main() {
	err := shim.Start(new())
	if err != nil {
		fmt.Printf("Error starting DocumentCC chaincode: %s", err)
	}
}
