/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package versionhandler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hyperledger/fabric/common/flogging"
	"github.com/trustbloc/sidetree-core-go/pkg/restapi/common"

	"github.com/trustbloc/sidetree-fabric/pkg/httpserver"
)

var logger = flogging.MustGetLogger("sidetree_peer")

const (
	versionParam = "version"
)

// Version handles version requests
type Version struct {
	path        string
	moduleName  string
	version     string
	channelID   string
	jsonMarshal func(v interface{}) ([]byte, error)
}

// New returns a new version handler
func New(channelID, basePath, module, version string) *Version {
	return &Version{
		path:        fmt.Sprintf("%s/%s", basePath, versionParam),
		moduleName:  module,
		version:     version,
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
	return h.getVersion
}

func (h *Version) getVersion(w http.ResponseWriter, _ *http.Request) {
	rw := httpserver.NewResponseWriter(w)

	resp := &Response{
		Name:    h.moduleName,
		Version: h.version,
	}

	versionBytes, err := h.jsonMarshal(resp)
	if err != nil {
		logger.Errorf("Unable to marshal CAS version: %s", err)

		rw.WriteError(httpserver.ServerError)
		return
	}

	logger.Debugf("[%s] ... returning version: %s", h.channelID, versionBytes)

	rw.Write(http.StatusOK, versionBytes, httpserver.ContentTypeJSON)
}
