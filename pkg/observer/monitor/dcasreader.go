/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package monitor

import (
	"github.com/trustbloc/fabric-peer-ext/pkg/collections/offledger/dcas/client"
	"github.com/trustbloc/sidetree-fabric/pkg/observer/common"
)

// SidetreeDCASReader reads DCAS data from the Sidetree DCAS store
type SidetreeDCASReader struct {
	dcasClientProvider common.DCASClientProvider
	channelID          string
}

// NewSidetreeDCASReader returns a SidetreeDCASReader
func NewSidetreeDCASReader(channelID string, dcasClientProvider common.DCASClientProvider) *SidetreeDCASReader {
	return &SidetreeDCASReader{
		channelID:          channelID,
		dcasClientProvider: dcasClientProvider,
	}
}

// Read returns the data for the given key from the Sidetree DCAS collection
func (r *SidetreeDCASReader) Read(key string) ([]byte, error) {
	dcasClient, err := r.dcasClient()
	if err != nil {
		return nil, err
	}
	return dcasClient.Get(common.SidetreeNs, common.SidetreeColl, key)
}

func (r *SidetreeDCASReader) dcasClient() (client.DCAS, error) {
	return r.dcasClientProvider.ForChannel(r.channelID)
}
