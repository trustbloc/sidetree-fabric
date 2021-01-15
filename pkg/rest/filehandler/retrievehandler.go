/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package filehandler

import (
	"bytes"
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
	ResolveDocument(idOrDocument string) (*document.ResolutionResult, error)
}

// Retrieve manages file retrievals from a DCAS store
type Retrieve struct {
	Config
	channelID    string
	resolver     documentResolver
	dcasProvider dcasClientProvider
}

type dcasClientProvider interface {
	GetDCASClient(channelID, namespace, coll string) (dcas.DCAS, error)
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

	logger.Debugf("[%s:%s:%s] Retrieving document for name [%s]", h.channelID, h.ChaincodeName, h.Collection, resourceName)

	fileBytes, contentType, err := h.doRetrieve(resourceName)
	if err != nil {
		writeError(rw, err.(*common.HTTPError).Status(), err)
		return
	}

	logger.Debugf("[%s:%s:%s] ... retrieved file [%s]: %s", h.channelID, h.ChaincodeName, h.Collection, resourceName, fileBytes)
	writeResponse(rw, http.StatusOK, fileBytes, contentType)
}

func (h *Retrieve) doRetrieve(resourceName string) ([]byte, string, error) {
	if resourceName == "" {
		return nil, "", common.NewHTTPError(http.StatusBadRequest, errors.New("resource name not provided"))
	}

	logger.Debugf("[%s:%s:%s] Resolving index file [%s]", h.channelID, h.ChaincodeName, h.Collection, h.IndexDocID)

	fileIndex, err := h.retrieveIndexDoc()
	if err != nil {
		return nil, "", err
	}

	cID := fileIndex.Mappings[resourceName]
	if cID == "" {
		logger.Debugf("[%s:%s:%s] Resource [%s] not found in index file", h.channelID, h.ChaincodeName, h.Collection, resourceName)
		return nil, "", common.NewHTTPError(http.StatusNotFound, errors.New(fileNotFound))
	}

	logger.Debugf("[%s:%s:%s] Got CID for [%s]: %s", h.channelID, h.ChaincodeName, h.Collection, resourceName, cID)

	fileBytes, contentType, err := h.retrieveFile(cID)
	if err != nil {
		return nil, "", err
	}

	if len(fileBytes) == 0 {
		logger.Debugf("[%s:%s:%s] CID [%s] not found in DCAS", h.channelID, h.ChaincodeName, h.Collection, cID)
		return nil, "", common.NewHTTPError(http.StatusNotFound, errors.New(fileNotFound))
	}

	return fileBytes, contentType, nil
}

func (h *Retrieve) retrieveIndexDoc() (*FileIndex, error) {
	logger.Debugf("[%s:%s:%s] Retrieving index document [%s]", h.channelID, h.ChaincodeName, h.Collection, h.IndexDocID)

	result, err := h.resolver.ResolveDocument(h.IndexDocID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			logger.Warnf("[%s:%s:%s] File index document not found in document store: [%s]", h.channelID, h.ChaincodeName, h.Collection, h.IndexDocID)
			return nil, common.NewHTTPError(http.StatusNotFound, errors.New("file index document not found"))
		}

		return nil, common.NewHTTPError(http.StatusInternalServerError, err)
	}

	if isDeactivated(result) {
		return nil, common.NewHTTPError(http.StatusGone, errors.New("document is no longer available"))
	}

	docBytes, err := json.Marshal(result.Document)
	if err != nil {
		logger.Errorf("[%s:%s:%s] Error marshalling file index document [%s]: %s", h.channelID, h.ChaincodeName, h.Collection, h.IndexDocID, err)
		return nil, common.NewHTTPError(http.StatusInternalServerError, errors.New(serverError))
	}

	logger.Debugf("[%s:%s:%s] Got index indexDoc: %s", h.channelID, h.ChaincodeName, h.Collection, docBytes)

	fileIndexDoc := &FileIndexDoc{}
	err = json.Unmarshal(docBytes, fileIndexDoc)
	if err != nil {
		logger.Errorf("[%s:%s:%s] Error unmarshalling file index document [%s]: %s", h.channelID, h.ChaincodeName, h.Collection, h.IndexDocID, err)
		return nil, common.NewHTTPError(http.StatusInternalServerError, errors.New(serverError))
	}

	if fileIndexDoc.ID != h.IndexDocID {
		logger.Errorf("[%s:%s:%s] Invalid file index payload: id [%s] in file index doc [%s] does not match [%s]", h.channelID, h.ChaincodeName, h.Collection, fileIndexDoc.ID, h.IndexDocID)
		return nil, common.NewHTTPError(http.StatusInternalServerError, errors.New(serverError))
	}

	if fileIndexDoc.FileIndex.BasePath != h.BasePath {
		logger.Errorf("[%s:%s:%s] Invalid file index payload: basePath [%s] in file index doc [%s] does not match base path of this endpoint [%s]", h.channelID, h.ChaincodeName, h.Collection, fileIndexDoc.FileIndex.BasePath, h.IndexDocID, h.BasePath)
		return nil, common.NewHTTPError(http.StatusInternalServerError, errors.New(serverError))
	}

	return &fileIndexDoc.FileIndex, nil
}

func (h *Retrieve) retrieveFile(cID string) ([]byte, string, error) {
	dcasClient, err := h.dcasProvider.GetDCASClient(h.channelID, h.ChaincodeName, h.Collection)
	if err != nil {
		logger.Errorf("[%s:%s:%s] Could not get DCAS client: %s", h.channelID, h.ChaincodeName, h.Collection, err)
		return nil, "", common.NewHTTPError(http.StatusInternalServerError, errors.New(serverError))
	}

	content := bytes.NewBuffer(nil)

	err = dcasClient.Get(cID, content)
	if err != nil {
		logger.Errorf("[%s:%s:%s] Error retrieving DCAS document for CID [%s]: %s", h.channelID, h.ChaincodeName, h.Collection, cID, err)
		return nil, "", common.NewHTTPError(http.StatusInternalServerError, errors.New(serverError))
	}

	if len(content.Bytes()) == 0 {
		logger.Debugf("[%s:%s:%s] File not found in DCAS for CID [%s]", h.channelID, h.ChaincodeName, h.Collection, cID)
		return nil, "", common.NewHTTPError(http.StatusNotFound, errors.New(fileNotFound))
	}

	f := &File{}
	if err := json.Unmarshal(content.Bytes(), f); err != nil {
		logger.Errorf("[%s:%s:%s] Error unmarshalling data for CID [%s]: %s", h.channelID, h.ChaincodeName, h.Collection, cID, err)
		return nil, "", common.NewHTTPError(http.StatusInternalServerError, errors.New(serverError))
	}

	if f.ContentType == "" {
		logger.Errorf("[%s:%s:%s] Content-type missing from file retrieved from DCAS for cID [%s]", h.channelID, h.ChaincodeName, h.Collection, cID)
		return nil, "", common.NewHTTPError(http.StatusInternalServerError, errors.New(serverError))
	}

	return f.Content, f.ContentType, nil
}

func isDeactivated(resolutionResult *document.ResolutionResult) bool {
	deactivated, ok := resolutionResult.DocumentMetadata[document.DeactivatedProperty]
	if !ok {
		return false
	}

	return deactivated.(bool)
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
