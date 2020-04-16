/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package blockchainhandler

import (
	"github.com/hyperledger/fabric/common/metadata"

	"github.com/trustbloc/sidetree-fabric/pkg/rest/versionhandler"
)

const (
	moduleName = "Hyperledger Fabric"
)

// Version handles version requests
type Version struct {
	*versionhandler.Version
}

// NewVersionHandler returns a new Version handler
func NewVersionHandler(channelID string, cfg Config) *Version {
	return &Version{
		Version: versionhandler.New(channelID, cfg.BasePath, moduleName, metadata.Version),
	}
}
