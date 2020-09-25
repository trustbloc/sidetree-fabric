/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package sidetreehandler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/hyperledger/fabric/common/flogging"
	"github.com/trustbloc/sidetree-core-go/pkg/api/protocol"
	"github.com/trustbloc/sidetree-core-go/pkg/restapi/common"

	"github.com/trustbloc/sidetree-fabric/pkg/httpserver"
)

var logger = flogging.MustGetLogger("sidetree_peer")

const (
	timeParam = "time"
)

type jsonMarshaller func(v interface{}) ([]byte, error)

type response struct {
	Version string `json:"version"`
	protocol.Protocol
}

// Version retrieves the current version and version parameters or the version and version parameters for a given transaction time
type Version struct {
	Config
	protocolClient protocol.Client
	channelID      string
	path           string
	jsonMarshal    jsonMarshaller
}

// NewVersionHandler returns a new Version handler
func NewVersionHandler(channelID string, cfg Config, pc protocol.Client) *Version {
	return &Version{
		Config:         cfg,
		protocolClient: pc,
		channelID:      channelID,
		path:           fmt.Sprintf("%s/version", cfg.BasePath),
		jsonMarshal:    json.Marshal,
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

func (h *Version) version(rw http.ResponseWriter, req *http.Request) {
	w := httpserver.NewResponseWriter(rw)

	pv, err := h.getProtocolVersion(req)
	if err != nil {
		logger.Debugf("[%s] Error getting version version: %s", h.channelID, err)

		w.WriteError(err)
		return
	}

	resp := response{
		Version:  pv.Version(),
		Protocol: pv.Protocol(),
	}

	pBytes, err := h.jsonMarshal(resp)
	if err != nil {
		logger.Errorf("[%s] Error marshalling resp: %s", h.channelID, err)

		w.WriteError(err)
		return
	}

	w.Write(http.StatusOK, pBytes, httpserver.ContentTypeJSON)
}

func (h *Version) getProtocolVersion(req *http.Request) (protocol.Version, error) {
	txnTime := getTime(req)
	if txnTime != "" {
		return h.getProtocolVersionAtTime(txnTime)
	}

	return h.getCurrentProtocolVersion()
}

func (h *Version) getCurrentProtocolVersion() (protocol.Version, error) {
	pv, err := h.protocolClient.Current()
	if err != nil {
		logger.Errorf("[%s] Error getting current version: %s", h.channelID, err)

		return nil, httpserver.ServerError
	}

	logger.Debugf("[%s] Current version: %+v", h.channelID, pv.Protocol())

	return pv, nil
}

func (h *Version) getProtocolVersionAtTime(txnTimeStr string) (protocol.Version, error) {
	txnTime, err := strconv.ParseUint(txnTimeStr, 10, 64)
	if err != nil {
		logger.Debugf("[%s] Invalid time [%s]: %s", h.channelID, txnTimeStr, err)

		return nil, httpserver.NewError(http.StatusBadRequest, fmt.Sprintf("invalid time: %s", txnTimeStr))
	}

	pv, err := h.protocolClient.Get(txnTime)
	if err != nil {
		if strings.Contains(err.Error(), "version parameters are not defined for blockchain time") {
			return nil, httpserver.NewError(http.StatusNotFound, fmt.Sprintf("version not found for time: %s", txnTimeStr))
		}

		logger.Errorf("[%s] Error getting version for time [%s]: %s", h.channelID, txnTimeStr, err)

		return nil, httpserver.ServerError
	}

	logger.Debugf("[%s] Version at time [%d]: %+v", h.channelID, txnTime, pv.Protocol())

	return pv, nil
}

func getTime(req *http.Request) string {
	values, ok := getParams(req)[timeParam]
	if !ok || len(values) == 0 {
		return ""
	}

	return values[0]
}

var getParams = func(req *http.Request) map[string][]string {
	return req.URL.Query()
}
