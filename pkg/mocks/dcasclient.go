/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"io"
	"io/ioutil"

	"github.com/ipfs/go-cid"
	mh "github.com/multiformats/go-multihash"
	"github.com/trustbloc/fabric-peer-ext/pkg/collections/offledger/dcas"
	dcasclient "github.com/trustbloc/fabric-peer-ext/pkg/collections/offledger/dcas/client"
)

// MockDCASClient implements a mock DCAS client
type MockDCASClient struct {
	data               map[string][]byte
	nodes              map[string]*dcasclient.Node
	errPut             error
	errGet, errGetNode error
	panicMessage       string
}

// NewDCASClient creates a mock DCAS client
func NewDCASClient() *MockDCASClient {
	return &MockDCASClient{
		data:  make(map[string][]byte),
		nodes: make(map[string]*dcasclient.Node),
	}
}

// WithData sets the key-value
func (m *MockDCASClient) WithData(key string, value []byte) *MockDCASClient {
	m.data[key] = value

	return m
}

// WithNode sets the key-node
func (m *MockDCASClient) WithNode(key string, node *dcasclient.Node) *MockDCASClient {
	m.nodes[key] = node

	return m
}

// WithPutError sets an error for Put operations
func (m *MockDCASClient) WithPutError(err error) *MockDCASClient {
	m.errPut = err

	return m
}

// WithGetError sets an error for Get operations
func (m *MockDCASClient) WithGetError(err error) *MockDCASClient {
	m.errGet = err

	return m
}

// WithGetNodeError sets an error for GetNode operations
func (m *MockDCASClient) WithGetNodeError(err error) *MockDCASClient {
	m.errGetNode = err

	return m
}

// WithPanic if set to true indicates that invoking any operation should cause a panic
func (m *MockDCASClient) WithPanic(msg string) *MockDCASClient {
	m.panicMessage = msg

	return m
}

// Put puts the data and returns the content ID for the value
func (m *MockDCASClient) Put(data io.Reader, opts ...dcasclient.Option) (string, error) {
	if m.errPut != nil {
		return "", m.errPut
	}

	if m.panicMessage != "" {
		panic(m.panicMessage)
	}

	value, err := ioutil.ReadAll(data)
	if err != nil {
		return "", err
	}

	cID, err := dcas.GetCID(value, dcas.CIDV1, cid.Raw, mh.SHA2_256)
	if err != nil {
		return "", err
	}

	m.data[cID] = value

	return cID, nil
}

// Delete deletes the values for the given content IDs.
func (m *MockDCASClient) Delete(...string) error {
	panic("not implemented")
}

// Get retrieves the value for the given content ID.
func (m *MockDCASClient) Get(cid string, w io.Writer) error {
	if m.errGet != nil {
		return m.errGet
	}

	if m.panicMessage != "" {
		panic(m.panicMessage)
	}

	value, ok := m.data[cid]
	if !ok {
		return nil
	}

	_, err := w.Write(value)
	return err
}

// GetNode retrieves the CAS Node for the given content ID. A node contains data and/or links to other nodes.
func (m *MockDCASClient) GetNode(cid string) (*dcasclient.Node, error) {
	if m.errGetNode != nil {
		return nil, m.errGetNode
	}

	if m.panicMessage != "" {
		panic(m.panicMessage)
	}

	return m.nodes[cid], nil
}
