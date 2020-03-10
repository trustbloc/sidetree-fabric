/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package filehandler

import (
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"

	"github.com/trustbloc/sidetree-core-go/pkg/restapi/common"
)

const (
	badRequest = "bad request"
)

// Upload manages file uploads to a DCAS store
type Upload struct {
	Config
	channelID    string
	dcasProvider dcasClientProvider
}

// NewUploadHandler returns a new file upload handler
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

// Handler returns the handler
func (h *Upload) Handler() common.HTTPRequestHandler {
	return h.upload
}

func (h *Upload) upload(rw http.ResponseWriter, req *http.Request) {
	request := &File{}
	if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
		logger.Debugf("Error unmarshalling request: %s", err)
		common.WriteError(rw, http.StatusBadRequest, errors.New(badRequest))
		return
	}

	id, err := h.doUpload(request)
	if err != nil {
		common.WriteError(rw, err.(*common.HTTPError).Status(), err)
		return
	}

	common.WriteResponse(rw, http.StatusOK, id)
}

// doUpload uploads the file and returns the CAS key of the file.
func (h *Upload) doUpload(request *File) (string, error) {
	if request.ContentType == "" {
		return "", common.NewHTTPError(http.StatusBadRequest, errors.New("content type is required"))
	}

	if len(request.Content) == 0 {
		return "", common.NewHTTPError(http.StatusBadRequest, errors.New("content is required"))
	}

	bytes, err := json.Marshal(request)
	if err != nil {
		logger.Errorf("[%s] Could not marshal data: %s", h.channelID, err)
		return "", common.NewHTTPError(http.StatusInternalServerError, errors.New(serverError))
	}

	client, err := h.dcasProvider.ForChannel(h.channelID)
	if err != nil {
		logger.Errorf("[%s] Could not get DCAS client: %s", h.channelID, err)
		return "", common.NewHTTPError(http.StatusInternalServerError, errors.New(serverError))
	}

	key, err := client.Put(h.ChaincodeName, h.Collection, bytes)
	if err != nil {
		logger.Errorf("[%s] Error storing file to DCAS collection [%s:%s]: %s", h.channelID, h.ChaincodeName, h.Collection, err)
		return "", common.NewHTTPError(http.StatusInternalServerError, errors.New(serverError))
	}

	logger.Debugf("[%s] Successfully uploaded file to DCAS [%s:%s:%s]", h.channelID, h.ChaincodeName, h.Collection, key)

	return key, nil
}
