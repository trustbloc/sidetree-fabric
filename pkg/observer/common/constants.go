/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package common

const (
	// AnchorPrefix is the prefix that is that is used to persist anchors
	AnchorPrefix = "sidetreeanchor_"
)

// TxnInfo contains info that gets recorded on blockchain as part of Sidetree transaction
type TxnInfo struct {
	AnchorString string `json:"anchorString"`
	Namespace    string `json:"namespace"`
	// ProtocolGenesisTime is the genesis time of the protocol that was used to create the anchor
	ProtocolGenesisTime uint64 `json:"protocolGenesisTime"`
}
