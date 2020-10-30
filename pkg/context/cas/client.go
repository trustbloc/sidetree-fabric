/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package cas

import (
	"bytes"

	"github.com/pkg/errors"
	"github.com/trustbloc/fabric-peer-ext/pkg/collections/offledger/dcas/client"

	"github.com/trustbloc/sidetree-fabric/pkg/common/transienterr"
)

// Client implements client for accessing the underlying content addressable storage
type Client struct {
	dcasClient client.DCAS
}

// New returns a new CAS client
func New(dcasClient client.DCAS) *Client {
	return &Client{
		dcasClient: dcasClient,
	}
}

// Write writes the given content to content addressable storage
// returns the content ID (CID) of the content.
func (c *Client) Write(content []byte) (string, error) {
	key, err := c.dcasClient.Put(bytes.NewReader(content))
	if err != nil {
		return "", transienterr.New(err, transienterr.CodeDB)
	}

	return key, nil
}

// Read reads the content at the given address from content addressable storage
// returns the content of the given address.
func (c *Client) Read(address string) ([]byte, error) {
	b := bytes.NewBuffer(nil)

	err := c.dcasClient.Get(address, b)
	if err != nil {
		return nil, transienterr.New(err, transienterr.CodeDB)
	}

	if len(b.Bytes()) == 0 {
		return nil, transienterr.New(errors.Errorf("content not found for key [%s]", address), transienterr.CodeNotFound)
	}

	return b.Bytes(), nil
}
