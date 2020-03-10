/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package filehandler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/hyperledger/fabric/common/flogging"
	"github.com/pkg/errors"

	dcas "github.com/trustbloc/fabric-peer-ext/pkg/collections/offledger/dcas/client"

	"github.com/trustbloc/sidetree-core-go/pkg/document"
	"github.com/trustbloc/sidetree-core-go/pkg/restapi/common"
)

var logger = flogging.MustGetLogger("sidetree_peer")

const (
	serverError  = "server error"
	fileNotFound = "file not found"
)

type documentResolver interface {
	ResolveDocument(idOrDocument string) (document.Document, error)
}

// Retrieve manages file retrievals from a DCAS store
type Retrieve struct {
	Config
	channelID    string
	resolver     documentResolver
	dcasProvider dcasClientProvider
}

type dcasClientProvider interface {
	ForChannel(channelID string) (dcas.DCAS, error)
}

// NewRetrieveHandler returns a new file retrieve handler
func NewRetrieveHandler(channelID string, cfg Config, resolver documentResolver, dcasProvider dcasClientProvider) *Retrieve {
	return &Retrieve{
		Config:       cfg,
		resolver:     resolver,
		dcasProvider: dcasProvider,
		channelID:    channelID,
	}
}

// Path returns the context path
func (h *Retrieve) Path() string {
	return h.BasePath + "/{resourceName}"
}

// Method returns the HTTP method
func (h *Retrieve) Method() string {
	return http.MethodGet
}

// Handler returns the handler
func (h *Retrieve) Handler() common.HTTPRequestHandler {
	return h.retrieve
}

// retrieve retrieves a resource by name. First the file index Sidetree document for the path
// is retrieved and then the ID of the resource is looked up from the index document. The file
// is then retrieved from the DCAS store using that ID.
func (h *Retrieve) retrieve(rw http.ResponseWriter, req *http.Request) {
	resourceName := getResourceName(req)

	logger.Debugf("Retrieving document for name [%s]", resourceName)

	fileBytes, contentType, err := h.doRetrieve(resourceName)
	if err != nil {
		writeError(rw, err.(*common.HTTPError).Status(), err)
		return
	}

	logger.Debugf("... retrieved file [%s]: %s", resourceName, fileBytes)
	writeResponse(rw, http.StatusOK, fileBytes, contentType)
}

func (h *Retrieve) doRetrieve(resourceName string) ([]byte, string, error) {
	if resourceName == "" {
		return nil, "", common.NewHTTPError(http.StatusBadRequest, errors.New("resource name not provided"))
	}

	logger.Debugf("Resolving index file [%s]", h.IndexDocID)

	indexDoc, err := h.retrieveIndexDoc()
	if err != nil {
		return nil, "", err
	}

	logger.Debugf("Got index indexDoc: %+v", indexDoc)

	key := indexDoc.GetStringValue(resourceName)
	if key == "" {
		logger.Debugf("Resource [%s] not found in index file", resourceName)
		return nil, "", common.NewHTTPError(http.StatusNotFound, errors.New(fileNotFound))
	}

	logger.Debugf("Got CAS key for [%s]: %s", resourceName, key)

	fileBytes, contentType, err := h.retrieveFile(key)
	if err != nil {
		return nil, "", err
	}

	if len(fileBytes) == 0 {
		logger.Debugf("[%s] Key [%s:%s:%s] not found in DCAS", h.channelID, h.ChaincodeName, h.Collection, key)
		return nil, "", common.NewHTTPError(http.StatusNotFound, errors.New(fileNotFound))
	}

	return fileBytes, contentType, nil
}

func (h *Retrieve) retrieveIndexDoc() (document.Document, error) {
	logger.Debugf("Retrieving index document [%s]", h.IndexDocID)

	indexDoc, err := h.resolver.ResolveDocument(h.IndexDocID)
	if err == nil {
		return indexDoc, nil
	}

	if strings.Contains(err.Error(), "not found") {
		logger.Warnf("File index document not found in document store: [%s]", h.IndexDocID)
		return nil, common.NewHTTPError(http.StatusNotFound, errors.New("file index document not found"))
	}

	if strings.Contains(err.Error(), "was deleted") {
		return nil, common.NewHTTPError(http.StatusGone, errors.New("document is no longer available"))
	}

	return nil, common.NewHTTPError(http.StatusInternalServerError, err)
}

func (h *Retrieve) retrieveFile(key string) ([]byte, string, error) {
	dcasClient, err := h.dcasProvider.ForChannel(h.channelID)
	if err != nil {
		logger.Errorf("[%s] Could not get DCAS client: %s", h.channelID, err)
		return nil, "", common.NewHTTPError(http.StatusInternalServerError, errors.New(serverError))
	}

	value, err := dcasClient.Get(h.ChaincodeName, h.Collection, key)
	if err != nil {
		logger.Errorf("[%s] Error retrieving DCAS document for key [%s:%s:%s]: %s", h.channelID, h.ChaincodeName, h.Collection, key, err)
		return nil, "", common.NewHTTPError(http.StatusInternalServerError, errors.New(serverError))
	}

	if len(value) == 0 {
		logger.Debugf("[%s] File not found in DCAS for key [%s:%s:%s]", h.channelID, h.ChaincodeName, h.Collection, key)
		return nil, "", common.NewHTTPError(http.StatusNotFound, errors.New(fileNotFound))
	}

	f := &File{}
	if err := json.Unmarshal(value, f); err != nil {
		logger.Errorf("[%s] Error unmarshalling data for key [%s:%s:%s]: %s", h.channelID, h.ChaincodeName, h.Collection, key, err)
		return nil, "", common.NewHTTPError(http.StatusInternalServerError, errors.New(serverError))
	}

	if f.ContentType == "" {
		logger.Errorf("[%s] Content-type missing from file retrieved from DCAS for key [%s:%s:%s]", h.channelID, h.ChaincodeName, h.Collection, key)
		return nil, "", common.NewHTTPError(http.StatusInternalServerError, errors.New(serverError))
	}

	return f.Content, f.ContentType, nil
}

var getResourceName = func(req *http.Request) string {
	return mux.Vars(req)["resourceName"]
}

func writeResponse(rw http.ResponseWriter, status int, bytes []byte, contentType string) {
	rw.Header().Set("Content-Type", contentType)
	rw.WriteHeader(status)
	_, e := rw.Write(bytes)
	if e != nil {
		logger.Errorf("Unable to write response: %s", e)
	}
}

func writeError(rw http.ResponseWriter, status int, err error) {
	rw.Header().Set("Content-Type", "text/plain")
	rw.WriteHeader(status)
	_, e := rw.Write([]byte(err.Error()))
	if e != nil {
		logger.Errorf("Unable to write response: %s", e)
	}
}
