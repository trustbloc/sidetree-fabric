/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package blockchainhandler

// ResultCode specifies the status string of a blockchain result
type ResultCode = string

const (
	// InvalidTxNumOrTimeHash indicates that the given transaction number or time hash is invalid
	InvalidTxNumOrTimeHash ResultCode = "invalid_transaction_number_or_time_hash"
	// InvalidFromTime indicates that the block time parameter is missing or invalid
	InvalidFromTime ResultCode = "invalid_from_time"
	// InvalidMaxBlocks indicates that the max blocks parameter is missing or invalid
	InvalidMaxBlocks ResultCode = "invalid_max_blocks"
	// InvalidTimeHash indicates that the time hash parameter is missing or invalid
	InvalidTimeHash ResultCode = "invalid_time_hash"
)

const (
	headerField       = "header"
	numberField       = "number"
	dataHashField     = "data_hash"
	previousHashField = "previous_Hash"
	dataField         = "data"
	metadataField     = "metadata"
)

const (
	fromTimeParam     = "from-time"
	maxBlocksParam    = "max-blocks"
	dataEncodingParam = "data-encoding"
)

// DataEncoding specifies the encoding of the data sections of the blocks in the response
type DataEncoding string

const (
	// DataEncodingBase64 indicates that the data sections of the returned block should be encoded using Base64 (standard) encoding
	DataEncodingBase64 DataEncoding = "base64"
	// DataEncodingBase64URL indicates that the data sections of the returned block should be encoded using Base64 URL-encoding
	DataEncodingBase64URL DataEncoding = "base64url"
	// DataEncodingJSON indicates that the data sections of the returned block should be returned in JSON format
	DataEncodingJSON DataEncoding = "json"
)

// TimeResponse contains the response from the /time request
type TimeResponse struct {
	Time string `json:"time"`
	// Hash is the base64 URL-encoded hash of the block.
	Hash string `json:"hash"`
	// PreviousHash is the base64 URL-encoded hash of the previous block's header.
	// This value may be used as the hash value to retrieve the previous block
	// using the '/time/{hash}' endpoint.
	PreviousHash string `json:"previous_hash"`
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

// Block is the JSON object representation of a block
type Block map[string]interface{}

// BlockHeader contains a block header.
type BlockHeader struct {
	Number       uint64 `json:"number"`
	DataHash     string `json:"data_hash"`
	PreviousHash string `json:"previous_hash"`
}

// BlockResponse contains a block.
type BlockResponse struct {
	Header   BlockHeader `json:"header"`
	Data     string      `json:"data"`
	MetaData string      `json:"metadata"`
}
