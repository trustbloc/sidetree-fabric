/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"container/list"
	"fmt"
	"strings"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"github.com/pkg/errors"

	"github.com/op/go-logging"
)

// Logger for the shim package.
var mockLogger = logging.MustGetLogger("mock")

/*
	The following code is taken from chaincode.go
*/

type stateMap map[string][]byte

// MockStub is an implementation of ChaincodeStubInterface for unit testing chaincode.
// Use this instead of ChaincodeStub in your chaincode's unit test calls to Init or Invoke.
type MockStub struct {
	shim.MockStub
	// arguments the stub was called with
	args [][]byte
	// A pointer back to the chaincode that will invoke this, set by constructor.
	// If a peer calls this stub, the chaincode will be invoked from here.
	cc shim.Chaincode
	// registered list of other MockStub chaincodes that can be called from this MockStub
	Invokables map[string]*MockStub
	// Transient map
	Transient map[string][]byte

	//PvtState contains private state
	PvtState map[string]stateMap

	// Keys stores the list of mapped values in lexical order
	Keys map[string]*list.List

	// Errors used for testing
	GetPrivateErr error
	PutPrivateErr error
}

// GetTransient returns transient map
func (stub *MockStub) GetTransient() (map[string][]byte, error) {
	return stub.Transient, nil
}

//GetArgs returns args
func (stub *MockStub) GetArgs() [][]byte {
	return stub.args
}

// InvokeChaincode calls a peered chaincode.
// E.g. stub1.InvokeChaincode("stub2Hash", funcArgs, channel)
// Before calling this make sure to create another MockStub stub2, call stub2.MockInit(uuid, func, args)
// and register it with stub1 by calling stub1.MockPeerChaincode("stub2Hash", stub2)
func (stub *MockStub) InvokeChaincode(chaincodeName string, args [][]byte, channel string) pb.Response {

	// Internally we use chaincode name as a composite name
	if channel != "" {
		chaincodeName = chaincodeName + "/" + channel
	}
	otherStub := stub.Invokables[chaincodeName]
	mockLogger.Debug("MockStub", stub.Name, "Invoking peer chaincode", otherStub.Name, args)

	res := otherStub.MockInvoke(stub.TxID, args)
	mockLogger.Debug("MockStub", stub.Name, "Invoked peer chaincode", otherStub.Name, "got", fmt.Sprintf("%+v", res))
	return res
}

// GetState retrieves the value for a given key from the ledger
func (stub *MockStub) GetState(key string) ([]byte, error) {
	return stub.getState("", key)
}

// PutState writes the specified `value` and `key` into the ledger.
func (stub *MockStub) PutState(key string, value []byte) error {
	return stub.putState("", key, value)
}

// DelState removes the specified `key` and its value from the ledger.
func (stub *MockStub) DelState(key string) error {
	return stub.delState("", key)
}

func (stub *MockStub) getStateMap(collection string) stateMap {
	sm, ok := stub.PvtState[collection]
	if !ok {
		sm = make(map[string][]byte)
		stub.PvtState[collection] = sm
	}
	return sm
}

func (stub *MockStub) getState(collection, key string) ([]byte, error) {
	value := stub.getStateMap(collection)[key]
	mockLogger.Debug("MockStub", stub.Name, "Getting", collection, key, value)
	return value, nil
}

func (stub *MockStub) delState(collection, key string) error {
	sm := stub.getStateMap(collection)
	mockLogger.Debug("MockStub", stub.Name, "Deleting", collection, key, sm[key])

	delete(sm, key)

	keys := stub.getKeys(collection)
	for elem := keys.Front(); elem != nil; elem = elem.Next() {
		if strings.Compare(key, elem.Value.(string)) == 0 {
			keys.Remove(elem)
		}
	}

	return nil
}

func (stub *MockStub) getKeys(collection string) *list.List {
	if _, ok := stub.Keys[collection]; !ok {
		stub.Keys[collection] = list.New()
	}
	return stub.Keys[collection]
}

func (stub *MockStub) putState(collection, key string, value []byte) error {
	if stub.TxID == "" {
		err := errors.New("cannot PutState without a transactions - call stub.MockTransactionStart()?")
		mockLogger.Errorf("%+v", err)
		return err
	}

	mockLogger.Debug("MockStub", stub.Name, "Putting", key, value)
	stub.getStateMap(collection)[key] = value

	// insert key into ordered list of keys
	keys := stub.getKeys(collection)
	for elem := keys.Front(); elem != nil; elem = elem.Next() {
		elemValue := elem.Value.(string)
		comp := strings.Compare(key, elemValue)
		mockLogger.Debug("MockStub", stub.Name, "Compared", key, elemValue, " and got ", comp)
		if comp < 0 {
			// key < elem, insert it before elem
			keys.InsertBefore(key, elem)
			mockLogger.Debug("MockStub", stub.Name, "Key", key, " inserted before", elem.Value)
			break
		} else if comp == 0 {
			// keys exists, no need to change
			mockLogger.Debug("MockStub", stub.Name, "Key", key, "already in State")
			break
		} else { // comp > 0
			// key > elem, keep looking unless this is the end of the list
			if elem.Next() == nil {
				keys.PushBack(key)
				mockLogger.Debug("MockStub", stub.Name, "Key", key, "appended")
				break
			}
		}
	}

	// special case for empty Keys list
	if keys.Len() == 0 {
		keys.PushFront(key)
		mockLogger.Debug("MockStub", stub.Name, "Key", key, "is first element in list")
	}

	return nil
}

// GetPrivateData implements get private data
func (stub *MockStub) GetPrivateData(collection string, key string) ([]byte, error) {

	if stub.GetPrivateErr != nil {
		if stub.GetPrivateErr.Error() == "panic" {
			panic("test panic")
		}
		return nil, stub.GetPrivateErr
	}

	return stub.getState(collection, key)
}

// PutPrivateData implements put private data
func (stub *MockStub) PutPrivateData(collection string, key string, value []byte) error {

	if stub.PutPrivateErr != nil {
		if stub.PutPrivateErr.Error() == "panic" {
			panic("test panic")
		}
		return stub.PutPrivateErr
	}

	return stub.putState(collection, key, value)
}

//NewMockStub initializes state map, private state, transients, invokables etc.
func NewMockStub(name string, cc shim.Chaincode) *MockStub {
	mockLogger.Debug("MockStub(", name, cc, ")")
	s := new(MockStub)
	s.Name = name
	s.cc = cc
	s.State = make(map[string][]byte)
	s.Invokables = make(map[string]*MockStub)
	s.Keys = make(map[string]*list.List)
	s.Transient = make(map[string][]byte)
	s.PvtState = make(map[string]stateMap)

	return s
}

// MockInit calls init chaincode
func (stub *MockStub) MockInit(uuid string, args [][]byte) pb.Response {
	stub.args = args
	stub.MockTransactionStart(uuid)
	res := stub.cc.Init(stub)
	stub.MockTransactionEnd(uuid)
	return res
}

//MockInvoke invokes chaincode
func (stub *MockStub) MockInvoke(uuid string, args [][]byte) pb.Response {
	stub.args = args
	stub.MockTransactionStart(uuid)
	res := stub.cc.Invoke(stub)
	stub.MockTransactionEnd(uuid)
	return res
}
