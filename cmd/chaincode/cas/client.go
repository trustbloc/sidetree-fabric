/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package cas

import (
	"github.com/btcsuite/btcutil/base58"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/pkg/errors"
	"github.com/trustbloc/fabric-peer-ext/pkg/collections/offledger/dcas"
)

// New returns a new client for managing content
func New(stub shim.ChaincodeStubInterface, collection string) *Client {
	client := &Client{stub: stub, collection: collection}
	client.Init(stub)
	return client
}

// Init initializes the client
func (mc *Client) Init(stub shim.ChaincodeStubInterface) {
	mc.stub = stub
}

// Client implements writing and reading content
type Client struct {
	stub       shim.ChaincodeStubInterface
	collection string
}

// Write stores content to DCAS.
// returns the SHA256 hash in base64url encoding which represents the address of the content
func (mc *Client) Write(content []byte) (string, error) {

	address, bytes, err := dcas.GetCASKeyAndValue(content)
	if err != nil {
		return "", err
	}

	if err := mc.stub.PutPrivateData(mc.collection, base58.Encode([]byte(address)), bytes); err != nil {
		return "", errors.Wrap(err, "failed to store content")
	}

	return address, nil
}

// Read reads the content of the given address in DCAS.
// returns the content of the given address
func (mc *Client) Read(address string) ([]byte, error) {

	payload, err := mc.stub.GetPrivateData(mc.collection, base58.Encode([]byte(address)))
	if err != nil {
		return nil, errors.Wrap(err, "failed to read content")
	}

	return payload, nil
}

// Query performs a "rich" query against a given private collection.
// It is only supported for state databases that support rich query.
func (mc *Client) Query(query string) ([][]byte, error) {

	iter, err := mc.stub.GetPrivateDataQueryResult(mc.collection, query)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to query content for: %s", query)
	}

	var results [][]byte
	for iter.HasNext() {
		elem, err := iter.Next()
		if err != nil {
			return nil, errors.Wrap(err, "failed to retrieve key and value in the range")
		}
		results = append(results, elem.GetValue())
	}

	return results, nil
}
