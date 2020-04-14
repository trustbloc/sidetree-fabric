/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package dcashandler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/trustbloc/sidetree-core-go/pkg/restapi/common"
	"github.com/trustbloc/sidetree-fabric/pkg/httpserver"
)

const (
	moduleName   = "cas"
	versionParam = "version"
)

// Version handles version requests
type Version struct {
	Config
	path        string
	channelID   string
	jsonMarshal func(v interface{}) ([]byte, error)
}

// NewVersionHandler returns a new Version handler
func NewVersionHandler(channelID string, cfg Config) *Version {
	return &Version{
		Config:      cfg,
		path:        fmt.Sprintf("%s/%s", cfg.BasePath, versionParam),
		channelID:   channelID,
		jsonMarshal: json.Marshal,
	}
}

// Path returns the context path
func (h *Version) Path() string {
	return h.path
}

// Method returns the HTTP method
func (h *Version) Method() string {
	return http.MethodGet
}

// Handler returns the request handler
func (h *Version) Handler() common.HTTPRequestHandler {
	return h.version
}

func (h *Version) version(w http.ResponseWriter, _ *http.Request) {
	rw := httpserver.NewResponseWriter(w)

	resp := &VersionResponse{
		Name:    moduleName,
		Version: h.Version,
	}

	versionBytes, err := h.jsonMarshal(resp)
	if err != nil {
		logger.Errorf("Unable to marshal CAS version: %s", err)

		rw.WriteError(httpserver.ServerError)
		return
	}

	logger.Debugf("[%s:%s:%s] ... returning version: %s", h.channelID, h.ChaincodeName, h.Collection, versionBytes)

	rw.Write(http.StatusOK, versionBytes, httpserver.ContentTypeJSON)
}
