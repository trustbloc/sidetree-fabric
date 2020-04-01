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

	// Observer indicates that this node observes batch writes to the ledger and persists document operations to the DB
	Observer = "sidetree-observer"

	// Monitor indicates that this node periodically monitors the blocks and checks that all document operations have been persisted to the DB
	Monitor = "sidetree-monitor"

	// Resolver indicates that this node exposes a REST API to resolve documents from the document operation store
	Resolver = "sidetree-resolver"
)

// IsObserver returns true if this node has the Observer role
func IsObserver() bool {
	return roles.HasRole(Observer)
}

// IsMonitor returns true if this node has the Monitor
func IsMonitor() bool {
	return roles.HasRole(Monitor)
}

// IsResolver returns true if this node has the Resolver role
func IsResolver() bool {
	return roles.HasRole(Resolver)
}

// IsBatchWriter returns true if this node has the Batch-Writer role
func IsBatchWriter() bool {
	return roles.HasRole(BatchWriter)
}
