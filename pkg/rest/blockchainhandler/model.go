/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package blockchainhandler

// TimeResponse contains the response from the /time request
type TimeResponse struct {
	Time string `json:"time"`
	Hash string `json:"hash"`
}

// TransactionsResponse contains a set of transactions and a boolean that indicates
// whether or not there are more transactions available to return.
type TransactionsResponse struct {
	More         bool          `json:"moreTransactions"`
	Transactions []Transaction `json:"transactions"`
}

// Transaction contains data for a single Sidetree transaction
type Transaction struct {
	TransactionNumber   uint64 `json:"transactionNumber"`
	TransactionTime     uint64 `json:"transactionTime"`
	TransactionTimeHash string `json:"transactionTimeHash"`
	AnchorString        string `json:"anchorString"`
}
