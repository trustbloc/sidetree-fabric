/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package dcashandler

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/trustbloc/sidetree-core-go/pkg/restapi/common"

	"github.com/trustbloc/sidetree-fabric/pkg/httpserver"
)

// Upload manages uploads to a DCAS store
type Upload struct {
	Config
	channelID    string
	dcasProvider dcasClientProvider
}

// NewUploadHandler returns a new DCAS upload handler
func NewUploadHandler(channelID string, cfg Config, dcasProvider dcasClientProvider) *Upload {
	return &Upload{
		Config:       cfg,
		channelID:    channelID,
		dcasProvider: dcasProvider,
	}
}

// Path returns the context path
func (h *Upload) Path() string {
	return h.BasePath
}

// Method returns the HTTP method
func (h *Upload) Method() string {
	return http.MethodPost
}

// Handler returns the request handler
func (h *Upload) Handler() common.HTTPRequestHandler {
	return h.upload
}

// upload uploads the content and responds with the hash of the content.
func (h *Upload) upload(w http.ResponseWriter, req *http.Request) {
	rw := newUploadWriter(w)

	hash, err := h.doUpload(req.Body)
	if err != nil {
		rw.WriteError(err)
		return
	}

	rw.Write(hash)
}

func (h *Upload) doUpload(content io.Reader) (string, error) {
	client, err := h.dcasProvider.GetDCASClient(h.channelID, h.ChaincodeName, h.Collection)
	if err != nil {
		logger.Errorf("[%s:%s:%s] Could not get DCAS client: %s", h.channelID, h.ChaincodeName, h.Collection, err)

		return "", httpserver.ServerError
	}

	hash, err := client.Put(content)
	if err != nil {
		logger.Errorf("[%s:%s:%s] Error storing content to DCAS: %s", h.channelID, h.ChaincodeName, h.Collection, err)

		return "", httpserver.ServerError
	}

	logger.Debugf("[%s:%s:%s] Successfully uploaded content to DCAS with hash [%s]", h.channelID, h.ChaincodeName, h.Collection, hash)

	return hash, nil
}

type uploadWriter struct {
	*httpserver.ResponseWriter
	jsonMarshal func(v interface{}) ([]byte, error)
}

func newUploadWriter(rw http.ResponseWriter) *uploadWriter {
	return &uploadWriter{
		ResponseWriter: httpserver.NewResponseWriter(rw),
		jsonMarshal:    json.Marshal,
	}
}

func (rw *uploadWriter) Write(hash string) {
	resp := &UploadResponse{
		Hash: hash,
	}

	respBytes, e := rw.jsonMarshal(resp)
	if e != nil {
		logger.Errorf("Unable to marshal response: %s", e)

		rw.WriteError(httpserver.ServerError)
		return
	}

	rw.ResponseWriter.Write(http.StatusOK, respBytes, httpserver.ContentTypeJSON)
}
