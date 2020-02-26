/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package doc

import (
	"github.com/hyperledger/fabric-chaincode-go/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	ccapi "github.com/hyperledger/fabric/extensions/chaincode/api"
)

const (
	ccVersion = "v1"

	collection = "docs"

	couchDB       = "couchdb"
	docsCollIndex = `{"index": {"fields": ["id"]}, "ddoc": "indexIDDoc", "name": "indexID", "type": "json"}`
)

// DocumentCC is used to setup database, collection and indexes for documents
type DocumentCC struct {
	name string
}

// New returns chaincode
func New(name string) *DocumentCC {
	cc := &DocumentCC{
		name: name,
	}
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
	return shim.Error("invoke is not supported")
}
