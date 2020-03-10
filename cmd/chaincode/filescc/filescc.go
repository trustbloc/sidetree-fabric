/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package filescc

import (
	"github.com/hyperledger/fabric-chaincode-go/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"

	ccapi "github.com/hyperledger/fabric/extensions/chaincode/api"
)

const (
	v1 = "v1"
)

// FilesCC is the chaincode that stores files in DCAS
type FilesCC struct {
	name string
}

// New returns a new example chaincode instance
func New(name string) *FilesCC {
	return &FilesCC{
		name: name,
	}
}

// Name returns the name of this chaincode
func (cc *FilesCC) Name() string { return cc.name }

// Version returns the version of the chaincode
func (cc *FilesCC) Version() string { return v1 }

// Chaincode returns this chaincode
func (cc *FilesCC) Chaincode() shim.Chaincode { return cc }

// GetDBArtifacts returns Couch DB indexes (if applicable)
func (cc *FilesCC) GetDBArtifacts() map[string]*ccapi.DBArtifacts { return nil }

// Init is not used
func (cc *FilesCC) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

// Invoke not implemented
func (cc *FilesCC) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Error("Invoke not implemented")
}
