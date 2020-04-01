/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package monitor

import (
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/trustbloc/sidetree-fabric/pkg/observer/common"
)

const (
	// MetaDataColName is the name of the meta-data collection used by the Monitor
	// to store peer-specific information
	MetaDataColName = "meta_data"
)

var errMetaDataNotFound = errors.New("not found")

// MetaDataStore manages the persistence and retrieval of peer-specific meta data
type MetaDataStore struct {
	channelID         string
	peerID            string
	ChaincodeName     string
	offLedgerProvider common.OffLedgerClientProvider
}

// NewMetaDataStore returns a new meta data store
func NewMetaDataStore(channelID, peerID string, ccName string, offLedgerProvider common.OffLedgerClientProvider) *MetaDataStore {
	return &MetaDataStore{
		channelID:         channelID,
		peerID:            peerID,
		ChaincodeName:     ccName,
		offLedgerProvider: offLedgerProvider,
	}
}

// Get retrieves the meta-data for this peer
func (m *MetaDataStore) Get() (*MetaData, error) {
	client, err := m.offLedgerProvider.ForChannel(m.channelID)
	if err != nil {
		return nil, err
	}

	data, err := client.Get(m.ChaincodeName, MetaDataColName, m.peerID)
	if err != nil {
		return nil, errors.WithMessage(err, "error retrieving meta-data")
	}

	if len(data) == 0 {
		logger.Debugf("[%s] No MetaData exists for peer [%s]", m.channelID, m.peerID)
		return nil, errMetaDataNotFound
	}

	metaData := &MetaData{}
	err = json.Unmarshal(data, metaData)
	if err != nil {
		return nil, errors.WithMessage(err, "error unmarshalling meta-data")
	}

	return metaData, nil
}

// Put persists the meta data for this peer
func (m *MetaDataStore) Put(data *MetaData) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return errors.WithMessage(err, "error marshalling meta-data")
	}

	client, err := m.offLedgerProvider.ForChannel(m.channelID)
	if err != nil {
		return err
	}

	return client.Put(m.ChaincodeName, MetaDataColName, m.peerID, bytes)
}
