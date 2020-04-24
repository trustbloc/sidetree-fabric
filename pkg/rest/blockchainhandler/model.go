/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package blockchainhandler

// ResultCode specifies the status string of a blockchain result
type ResultCode = string

var (
	// InvalidTxNumOrTimeHash indicates that the given transaction number or time hash is invalid
	InvalidTxNumOrTimeHash ResultCode = "invalid_transaction_number_or_time_hash"
)

// TimeResponse contains the response from the /time request
type TimeResponse struct {
	Time string `json:"time"`
	Hash string `json:"hash"`
}

// TransactionsResponse contains a set of transactions and a boolean that indicates
// whether or not there are more transactions available to return.
type TransactionsResponse struct {
	More         bool          `json:"more_transactions"`
	Transactions []Transaction `json:"transactions"`
}

// Transaction contains data for a single Sidetree transaction
type Transaction struct {
	TransactionNumber   uint64 `json:"transaction_number"`
	TransactionTime     uint64 `json:"transaction_time"`
	TransactionTimeHash string `json:"transaction_time_hash"`
	AnchorString        string `json:"anchor_string"`
}

// ErrorResponse contains the error code for a failed response
type ErrorResponse struct {
	Code string `json:"code"`
}
