/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package common

import (
	"encoding/json"

	"github.com/trustbloc/fabric-peer-ext/pkg/collections/offledger/dcas"
)

// MarshalDCAS marshals the given object to a JSON format, normalizes the content
// (i.e. if the content is a JSON doc then the fields are marshaled in a deterministic order)
// and returns the content-addressable key (encoded in base64) along with the normalized value.
func MarshalDCAS(obj interface{}) (key string, value []byte, err error) {
	bytes, err := json.Marshal(obj)
	if err != nil {
		return "", nil, err
	}
	return dcas.GetCASKeyAndValue(bytes)
}
