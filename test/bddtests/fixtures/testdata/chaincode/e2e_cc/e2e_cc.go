/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"
)

type invokeFunc func(stub shim.ChaincodeStubInterface, args []string) pb.Response
type funcMap map[string]invokeFunc

const (
	getFunc = "get"
	putFunc = "put"
)

// ExampleCC example chaincode that puts and gets state and private data
type ExampleCC struct {
	funcRegistry funcMap
}

// Init ...
func (cc *ExampleCC) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

// Invoke invoke the chaincode with a given function
func (cc *ExampleCC) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Printf("########### example_cc Invoke ###########")
	function, args := stub.GetFunctionAndParameters()
	if function == "" {
		return shim.Error("Expecting function")
	}

	f, ok := cc.funcRegistry[function]
	if !ok {
		return shim.Error(fmt.Sprintf("Unknown function [%s]. Expecting one of: %v", function, cc.functions()))
	}

	return f(stub, args)
}

func (cc *ExampleCC) put(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 2 {
		return shim.Error("Invalid args. Expecting key and value")
	}

	key := args[0]
	value := args[1]

	existingValue, err := stub.GetState(key)
	if err != nil {
		return shim.Error(fmt.Sprintf("Error getting data for key [%s]: %s", key, err))
	}
	if existingValue != nil {
		value = string(existingValue) + "-" + value
	}

	if err := stub.PutState(key, []byte(value)); err != nil {
		return shim.Error(fmt.Sprintf("Error putting data for key [%s]: %s", key, err))
	}

	return shim.Success([]byte(value))
}

func (cc *ExampleCC) get(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("Invalid args. Expecting key")
	}

	key := args[0]

	value, err := stub.GetState(key)
	if err != nil {
		return shim.Error(fmt.Sprintf("Error getting data for key [%s]: %s", key, err))
	}

	return shim.Success([]byte(value))
}

func (cc *ExampleCC) initRegistry() {
	cc.funcRegistry = make(map[string]invokeFunc)
	cc.funcRegistry[getFunc] = cc.get
	cc.funcRegistry[putFunc] = cc.put
}

func (cc *ExampleCC) functions() []string {
	var funcs []string
	for key := range cc.funcRegistry {
		funcs = append(funcs, key)
	}
	return funcs
}

func main() {
	cc := new(ExampleCC)
	cc.initRegistry()
	err := shim.Start(cc)
	if err != nil {
		fmt.Printf("Error starting example chaincode: %s", err)
	}
}
