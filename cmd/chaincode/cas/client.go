/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package cas

import (
	"bytes"

	"github.com/pkg/errors"
	dcasclient "github.com/trustbloc/fabric-peer-ext/pkg/collections/offledger/dcas/client"
)

// New returns a new client for managing content
func New(casClient dcasclient.DCAS) *Client {
	client := &Client{
		casClient: casClient,
	}

	return client
}

// Client implements writing and reading content
type Client struct {
	casClient dcasclient.DCAS
}

// Write stores content to DCAS.
// returns the content ID (CID) in V1 format (https://docs.ipfs.io/concepts/content-addressing/#version-1-v1)
func (mc *Client) Write(content []byte) (string, error) {
	address, err := mc.casClient.Put(bytes.NewReader(content))
	if err != nil {
		return "", errors.Wrap(err, "failed to store content")
	}

	return address, nil
}

// Read reads the content of the given content ID (CID) in DCAS.
// returns the content for the given CID. The CID must be encoded in the proper format,
// according to https://docs.ipfs.io/concepts/content-addressing/#identifier-formats
func (mc *Client) Read(cID string) ([]byte, error) {
	payload := bytes.NewBuffer(nil)

	err := mc.casClient.Get(cID, payload)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read content")
	}

	return payload.Bytes(), nil
}
