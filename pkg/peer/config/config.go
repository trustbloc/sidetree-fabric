/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package config

import (
	"time"
)

const (
	// GlobalMSPID is used as the consortium-wide MSP ID (i.e. non org-specific)
	GlobalMSPID = "general"

	// ProtocolComponentName is the name of the Sidetree protocol config component
	ProtocolComponentName = "protocol"

	// SidetreeAppName is the '=name of the Sidetree config application
	SidetreeAppName = "sidetree"

	// SidetreeAppVersion is the version of the Sidetree config application
	SidetreeAppVersion = "1"
)

// Namespace holds Sidetree namespace config
type Namespace struct {
	Namespace string
	BasePath  string
}

// Monitor holds Sidetree monitor config
type Monitor struct {
	Period time.Duration
}

// SidetreePeer holds peer-specific Sidetree config
type SidetreePeer struct {
	Monitor    Monitor
	Namespaces []Namespace
}

// Sidetree holds general Sidetree configuration
type Sidetree struct {
	BatchWriterTimeout time.Duration
}
