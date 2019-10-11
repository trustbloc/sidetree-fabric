/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package monitor

import (
	"github.com/trustbloc/sidetree-fabric/pkg/client"
	"github.com/trustbloc/sidetree-fabric/pkg/observer/common"
)

// SidetreeDCASReader reads DCAS data from the Sidetree DCAS store
type SidetreeDCASReader struct {
	dcasClientProvider dcasClientProvider
	channelID          string
}

// NewSidetreeDCASReader returns a SidetreeDCASReader
func NewSidetreeDCASReader(channelID string, dcasClientProvider dcasClientProvider) *SidetreeDCASReader {
	return &SidetreeDCASReader{
		channelID:          channelID,
		dcasClientProvider: dcasClientProvider,
	}
}

// Read returns the data for the given key from the Sidetree DCAS collection
func (r *SidetreeDCASReader) Read(key string) ([]byte, error) {
	return r.dcasClient().Get(common.SidetreeNs, common.SidetreeColl, key)
}

func (r *SidetreeDCASReader) dcasClient() client.DCAS {
	return r.dcasClientProvider.ForChannel(r.channelID)
}
