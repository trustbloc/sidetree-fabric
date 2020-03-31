/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package file

import (
	"github.com/hyperledger/fabric-chaincode-go/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"

	ccapi "github.com/hyperledger/fabric/extensions/chaincode/api"
)

const (
	v1 = "v1"
)

// Chaincode is the chaincode that stores files in DCAS
type Chaincode struct {
	name string
}

// New returns a new example chaincode instance
func New(name string) *Chaincode {
	return &Chaincode{
		name: name,
	}
}

// Name returns the name of this chaincode
func (cc *Chaincode) Name() string { return cc.name }

// Version returns the version of the chaincode
func (cc *Chaincode) Version() string { return v1 }

// Chaincode returns this chaincode
func (cc *Chaincode) Chaincode() shim.Chaincode { return cc }

// GetDBArtifacts returns Couch DB indexes (if applicable)
func (cc *Chaincode) GetDBArtifacts([]string) map[string]*ccapi.DBArtifacts { return nil }

// Init is not used
func (cc *Chaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

// Invoke not implemented
func (cc *Chaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Error("Invoke not implemented")
}
