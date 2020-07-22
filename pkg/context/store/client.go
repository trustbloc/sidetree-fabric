/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package store

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/pkg/errors"

	"github.com/hyperledger/fabric-protos-go/ledger/queryresult"

	"github.com/hyperledger/fabric/common/flogging"
	commonledger "github.com/hyperledger/fabric/common/ledger"

	"github.com/trustbloc/sidetree-core-go/pkg/api/batch"
	"github.com/trustbloc/sidetree-fabric/pkg/common/transienterr"
)

const (
	queryByUniqueSuffixTemplate = `{"selector":{"uniqueSuffix":"%s"},"use_index":["_design/indexUniqueSuffixDoc","indexUniqueSuffix"],"fields":["uniqueSuffix","type","encodedDelta","signedData","encodedSuffixData","transactionTime","transactionNumber","operationIndex"]}`
)

var logger = flogging.MustGetLogger("sidetree_context")

type store interface {
	Query(query string) (commonledger.ResultsIterator, error)
	Put(value []byte) error
}

// Client implements client for accessing document operations
type Client struct {
	channelID string
	namespace string
	store     store
}

// NewClient returns a new operation store client
func NewClient(channelID, namespace string, s store) *Client {
	return &Client{
		channelID: channelID,
		namespace: namespace,
		store:     s,
	}
}

// Get retrieves all document operations for specified document ID
func (c *Client) Get(uniqueSuffix string) ([]*batch.AnchoredOperation, error) {
	logger.Debugf("[%s-%s] Querying for operations for ID [%s]", c.channelID, c.namespace, uniqueSuffix)

	iter, err := c.store.Query(fmt.Sprintf(queryByUniqueSuffixTemplate, uniqueSuffix))
	if err != nil {
		return nil, transienterr.New(errors.Wrap(err, "failed to query document operations"), transienterr.CodeDB)
	}

	var ops [][]byte
	for {
		next, err := iter.Next()
		if err != nil {
			return nil, transienterr.New(errors.Wrap(err, "failed to retrieve key and value in the range"), transienterr.CodeDB)
		}
		if next == nil {
			break
		}
		kv := next.(*queryresult.KV)
		ops = append(ops, kv.Value)
	}

	if len(ops) == 0 {
		logger.Debugf("[%s-%s] No operations found for ID [%s]", c.channelID, c.namespace, uniqueSuffix)
		return nil, errors.New("uniqueSuffix not found in the store")
	}

	logger.Debugf("[%s-%s] Found operations for ID [%s]: %s", c.channelID, c.namespace, uniqueSuffix, ops)

	return getOperations(ops)
}

// Put stores an operation
func (c *Client) Put(ops []*batch.AnchoredOperation) error {
	for _, op := range ops {
		bytes, err := json.Marshal(op)
		if err != nil {
			return errors.Wrapf(err, "json marshal for op failed")
		}

		logger.Debugf("[%s-%s] Storing operation %s", c.channelID, c.namespace, bytes)

		err = c.store.Put(bytes)
		if err != nil {
			return transienterr.New(err, transienterr.CodeDB)
		}
	}

	return nil
}

func getOperations(ops [][]byte) ([]*batch.AnchoredOperation, error) {
	var operations []*batch.AnchoredOperation
	for _, opBytes := range ops {
		var op batch.AnchoredOperation
		if err := json.Unmarshal(opBytes, &op); err != nil {
			return nil, errors.Wrapf(err, "failed to unmarshal operation")
		}
		operations = append(operations, &op)
	}

	return sortChronologically(operations), nil
}

func sortChronologically(operations []*batch.AnchoredOperation) []*batch.AnchoredOperation {
	if len(operations) <= 1 {
		return operations
	}

	sort.Slice(operations, func(i, j int) bool {
		if operations[i].TransactionTime == operations[j].TransactionTime {
			return operations[i].TransactionNumber < operations[j].TransactionNumber
		}
		return operations[i].TransactionTime < operations[j].TransactionTime
	})

	return operations
}
