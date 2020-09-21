/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package protocolversion

import (
	"github.com/trustbloc/sidetree-fabric/pkg/protocolversion/factoryregistry"
	v0_1 "github.com/trustbloc/sidetree-fabric/pkg/protocolversion/versions/v0_1/factory"
)

const (
	// V0_1 version 0.1
	V0_1 = "0.1"
)

// RegisterFactories registers all protocol version factories
func RegisterFactories() {
	factoryregistry.Register(V0_1, v0_1.New())
}
