/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package discoveryhandler

import (
	"encoding/json"
	"net/http"

	"github.com/trustbloc/sidetree-core-go/pkg/restapi/common"

	"github.com/trustbloc/sidetree-fabric/pkg/httpserver"
	"github.com/trustbloc/sidetree-fabric/pkg/peer/discovery"
)

// Discover is a handler that returns available service endpoints
type Discover struct {
	Config
	servicesProvider
	path        string
	channelID   string
	jsonMarshal func(v interface{}) ([]byte, error)
}

type servicesProvider interface {
	ServicesForChannel(channelID string) []discovery.Service
}

// New returns a new discovery handler
func New(channelID string, cfg Config, servicesProvider servicesProvider) *Discover {
	return newHandler(channelID, cfg, servicesProvider, json.Marshal)
}

func newHandler(channelID string, cfg Config, servicesProvider servicesProvider, jsonMarshal func(v interface{}) ([]byte, error)) *Discover {
	return &Discover{
		Config:           cfg,
		servicesProvider: servicesProvider,
		path:             cfg.BasePath,
		channelID:        channelID,
		jsonMarshal:      jsonMarshal,
	}
}

// Path returns the context path
func (h *Discover) Path() string {
	return h.path
}

// Method returns the HTTP method
func (h *Discover) Method() string {
	return http.MethodGet
}

// Handler returns the request handler
func (h *Discover) Handler() common.HTTPRequestHandler {
	return h.retrieve
}

// retrieve retrieves discovery information
func (h *Discover) retrieve(rw http.ResponseWriter, req *http.Request) {
	w := httpserver.NewResponseWriter(rw)

	services := Services(h.ServicesForChannel(h.channelID)).FilterByParams(getParams(req))

	respBytes, err := h.jsonMarshal(services)
	if err != nil {
		w.WriteError(err)
		return
	}

	w.Write(http.StatusOK, respBytes, httpserver.ContentTypeJSON)
}

var getParams = func(req *http.Request) map[string][]string {
	return req.URL.Query()
}
