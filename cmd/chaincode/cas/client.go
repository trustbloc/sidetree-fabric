/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package cas

import (
	"crypto"
	"encoding/base64"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/pkg/errors"
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

	address := calculateAddress(content)

	if err := mc.stub.PutPrivateData(mc.collection, address, content); err != nil {
		return "", errors.Wrap(err, "failed to store content")
	}

	return address, nil
}

// Read reads the content of the given address in DCAS.
// returns the content of the given address
func (mc *Client) Read(address string) ([]byte, error) {

	payload, err := mc.stub.GetPrivateData(mc.collection, address)
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

// getHash will compute the hash for the supplied bytes using SHA256
func getHash(bytes []byte) []byte {
	h := crypto.SHA256.New()
	// added no lint directive because there's no error from source code
	// error cannot be produced, checked google source
	h.Write(bytes) //nolint
	return h.Sum(nil)
}

func calculateAddress(content []byte) string {
	hash := getHash(content)
	buf := make([]byte, base64.URLEncoding.EncodedLen(len(hash)))
	base64.URLEncoding.Encode(buf, hash)

	return string(buf)
}
