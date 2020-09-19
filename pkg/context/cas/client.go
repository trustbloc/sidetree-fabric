/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package cas

import (
	"github.com/pkg/errors"
	"github.com/trustbloc/fabric-peer-ext/pkg/collections/offledger/dcas/client"

	"github.com/trustbloc/sidetree-fabric/pkg/common/transienterr"
	"github.com/trustbloc/sidetree-fabric/pkg/config"
)

// Client implements client for accessing the underlying content addressable storage
type Client struct {
	config.DCAS
	dcasClient client.DCAS
}

// New returns a new CAS client
func New(dcasCfg config.DCAS, dcasClient client.DCAS) *Client {
	return &Client{
		DCAS:       dcasCfg,
		dcasClient: dcasClient,
	}
}

// Write writes the given content to content addressable storage
// returns the SHA256 hash in base64url encoding which represents the address of the content.
func (c *Client) Write(content []byte) (string, error) {
	key, err := c.dcasClient.Put(c.ChaincodeName, c.Collection, content)
	if err != nil {
		return "", transienterr.New(err, transienterr.CodeDB)
	}

	return key, nil
}

// Read reads the content at the given address from content addressable storage
// returns the content of the given address.
func (c *Client) Read(address string) ([]byte, error) {
	data, err := c.dcasClient.Get(c.ChaincodeName, c.Collection, address)
	if err != nil {
		return nil, transienterr.New(err, transienterr.CodeDB)
	}

	if len(data) == 0 {
		return nil, transienterr.New(errors.Errorf("content not found for key [%s]", address), transienterr.CodeNotFound)
	}

	return data, nil
}
