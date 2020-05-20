/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package role

import (
	"github.com/trustbloc/fabric-peer-ext/pkg/roles"
)

const (
	// BatchWriter indicates that this node exposes a REST API to submit document updates and writes batch files to the leder
	BatchWriter = "sidetree-batch-writer"

	// Observer indicates that this node is an 'active' that observes batch writes to the ledger and persists document operations to the DB
	Observer = "sidetree-observer"

	// ObserverStandby indicates that this node is a 'standby' observer, which means that the node will take over processing if
	// all nodes with an active observer role are down.
	ObserverStandby = "sidetree-observer-standby"

	// Resolver indicates that this node exposes a REST API to resolve documents from the document operation store
	Resolver = "sidetree-resolver"
)

// IsObserver returns true if this node has either an active or standby observer role
func IsObserver() bool {
	return roles.HasRole(Observer) || roles.HasRole(ObserverStandby)
}

// IsResolver returns true if this node has the Resolver role
func IsResolver() bool {
	return roles.HasRole(Resolver)
}

// IsBatchWriter returns true if this node has the Batch-Writer role
func IsBatchWriter() bool {
	return roles.HasRole(BatchWriter)
}
