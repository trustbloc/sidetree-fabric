/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package sidetreehandler

import (
	"github.com/trustbloc/sidetree-fabric/pkg/rest/versionhandler"
)

const (
	moduleName = "Sidetree"
)

// Version handles version requests
type Version struct {
	*versionhandler.Version
}

// NewVersionHandler returns a new Version handler
func NewVersionHandler(channelID string, cfg Config) *Version {
	return &Version{
		Version: versionhandler.New(channelID, cfg.BasePath, moduleName, cfg.Version),
	}
}
