/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package observer

import (
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/trustbloc/fabric-peer-ext/pkg/roles"

	"github.com/trustbloc/sidetree-fabric/pkg/observer/common"
)

const (
	// MetaDataColName is the name of the metadata collection used by the Observer
	// to store peer-specific information
	MetaDataColName = "meta_data"
)

var errMetaDataNotFound = errors.New("not found")

// MetadataStore manages the persistence and retrieval of peer-specific metadata
type MetadataStore struct {
	channelID         string
	metadataKey       string
	chaincodeName     string
	offLedgerProvider common.OffLedgerClientProvider
}

// NewMetaDataStore returns a new meta data store
func NewMetaDataStore(channelID string, peerConfig peerConfig, ccName string, offLedgerProvider common.OffLedgerClientProvider) *MetadataStore {
	var metadataKey string
	if roles.IsClustered() {
		metadataKey = peerConfig.MSPID()
	} else {
		metadataKey = peerConfig.PeerID()
	}

	return &MetadataStore{
		channelID:         channelID,
		metadataKey:       metadataKey,
		chaincodeName:     ccName,
		offLedgerProvider: offLedgerProvider,
	}
}

// Get retrieves the meta-data for this peer
func (m *MetadataStore) Get() (*Metadata, error) {
	client, err := m.offLedgerProvider.ForChannel(m.channelID)
	if err != nil {
		return nil, err
	}

	data, err := client.Get(m.chaincodeName, MetaDataColName, m.metadataKey)
	if err != nil {
		return nil, errors.WithMessage(err, "error retrieving meta-data")
	}

	if len(data) == 0 {
		logger.Debugf("[%s] No metadata exists for [%s]", m.channelID, m.metadataKey)
		return nil, errMetaDataNotFound
	}

	metaData := &Metadata{}
	err = json.Unmarshal(data, metaData)
	if err != nil {
		return nil, errors.WithMessage(err, "error unmarshalling meta-data")
	}

	return metaData, nil
}

// Put persists the meta data for this peer
func (m *MetadataStore) Put(data *Metadata) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return errors.WithMessage(err, "error marshalling meta-data")
	}

	client, err := m.offLedgerProvider.ForChannel(m.channelID)
	if err != nil {
		return err
	}

	return client.Put(m.chaincodeName, MetaDataColName, m.metadataKey, bytes)
}
