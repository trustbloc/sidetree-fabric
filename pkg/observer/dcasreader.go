/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package observer

import (
	"github.com/pkg/errors"
	"github.com/trustbloc/fabric-peer-ext/pkg/collections/offledger/dcas/client"

	"github.com/trustbloc/sidetree-fabric/pkg/common/transienterr"
	"github.com/trustbloc/sidetree-fabric/pkg/config"
	"github.com/trustbloc/sidetree-fabric/pkg/observer/common"
)

// SidetreeDCASReader reads DCAS data from the Sidetree DCAS store
type SidetreeDCASReader struct {
	config.DCAS
	dcasClientProvider common.DCASClientProvider
	channelID          string
}

// NewSidetreeDCASReader returns a SidetreeDCASReader
func NewSidetreeDCASReader(channelID string, dcasCfg config.DCAS, dcasClientProvider common.DCASClientProvider) *SidetreeDCASReader {
	return &SidetreeDCASReader{
		DCAS:               dcasCfg,
		channelID:          channelID,
		dcasClientProvider: dcasClientProvider,
	}
}

// Read returns the data for the given key from the Sidetree DCAS collection
func (r *SidetreeDCASReader) Read(key string) ([]byte, error) {
	dcasClient, err := r.dcasClient()
	if err != nil {
		return nil, transienterr.New(err)
	}

	content, err := dcasClient.Get(r.ChaincodeName, r.Collection, key)
	if err != nil {
		return nil, transienterr.New(err)
	}

	if len(content) == 0 {
		return nil, errors.Errorf("content not found for key [%s]", key)
	}

	return content, nil
}

func (r *SidetreeDCASReader) dcasClient() (client.DCAS, error) {
	return r.dcasClientProvider.ForChannel(r.channelID)
}
