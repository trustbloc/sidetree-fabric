/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package store

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-protos-go/ledger/queryresult"
	"github.com/hyperledger/fabric/common/flogging"
	"github.com/pkg/errors"
	"github.com/trustbloc/fabric-peer-ext/pkg/collections/offledger/dcas/client"
	"github.com/trustbloc/sidetree-core-go/pkg/api/batch"
	"github.com/trustbloc/sidetree-core-go/pkg/docutil"
)

const (
	documentCC        = "document_cc"
	collection        = "docs"
	queryByIDTemplate = `{"selector":{"id":"%s"},"use_index":["_design/indexIDDoc","indexID"],"fields":["id","encodedPayload","hashAlgorithmInMultiHashCode","operationIndex","patch","signature","signingKeyID","transactionNumber","transactionTime","type","uniqueSuffix"]}`
)

var logger = flogging.MustGetLogger("sidetree_context")

type dcasClientProvider interface {
	ForChannel(channelID string) (client.DCAS, error)
}

// Client implements client for accessing document operations
type Client struct {
	channelID     string
	namespace     string
	storeProvider dcasClientProvider
}

// New returns a new store client
func New(channelID, namespace string, storeProvider dcasClientProvider) *Client {
	return &Client{
		channelID:     channelID,
		namespace:     namespace,
		storeProvider: storeProvider,
	}
}

// Get retrieves all document operations for specified document ID
func (c *Client) Get(uniqueSuffix string) ([]batch.Operation, error) {
	id := c.namespace + docutil.NamespaceDelimiter + uniqueSuffix
	logger.Debugf("get operations for doc[%s]", id)

	sp, err := c.storeProvider.ForChannel(c.channelID)
	if err != nil {
		return nil, err
	}

	iter, err := sp.Query(documentCC, collection, fmt.Sprintf(queryByIDTemplate, id))
	if err != nil {
		return nil, errors.Wrap(err, "failed to query document operations")
	}

	var ops [][]byte
	for {
		next, err := iter.Next()
		if err != nil {
			return nil, errors.Wrap(err, "failed to retrieve key and value in the range")
		}
		if next == nil {
			break
		}
		kv := next.(*queryresult.KV)
		ops = append(ops, kv.Value)
	}

	if len(ops) == 0 {
		return nil, errors.New("uniqueSuffix not found in the store")
	}

	return getOperations(ops)
}

// Put stores an operation
func (c *Client) Put(op batch.Operation) error {
	// TODO: Not sure where/if this is useds
	panic("not implemented")
}

func getOperations(ops [][]byte) ([]batch.Operation, error) {
	var operations []batch.Operation
	for _, opBytes := range ops {
		var op batch.Operation
		if err := json.Unmarshal(opBytes, &op); err != nil {
			return nil, errors.Wrapf(err, "failed to unmarshal operation")
		}
		operations = append(operations, op)
	}

	return operations, nil
}
