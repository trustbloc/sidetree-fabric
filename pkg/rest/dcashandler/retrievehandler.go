/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package dcashandler

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/hyperledger/fabric/common/flogging"
	dcas "github.com/trustbloc/fabric-peer-ext/pkg/collections/offledger/dcas/client"
	"github.com/trustbloc/sidetree-core-go/pkg/restapi/common"
	"github.com/trustbloc/sidetree-fabric/pkg/httpserver"
)

var logger = flogging.MustGetLogger("sidetree_peer")

const (
	hashParam    = "hash"
	maxSizeParam = "max-size"

	encodingErrMsg = "selected encoding not supported"
)

// Retrieve manages file retrievals from the DCAS store
type Retrieve struct {
	Config
	path         string
	channelID    string
	dcasProvider dcasClientProvider
}

type dcasClientProvider interface {
	GetDCASClient(channelID, namespace, coll string) (dcas.DCAS, error)
}

// NewRetrieveHandler returns a new Retrieve handler
func NewRetrieveHandler(channelID string, cfg Config, dcasProvider dcasClientProvider) *Retrieve {
	return &Retrieve{
		Config:       cfg,
		path:         fmt.Sprintf("%s/{%s}", cfg.BasePath, hashParam),
		dcasProvider: dcasProvider,
		channelID:    channelID,
	}
}

// Path returns the context path
func (h *Retrieve) Path() string {
	return h.path
}

// Method returns the HTTP method
func (h *Retrieve) Method() string {
	return http.MethodGet
}

// Handler returns the request handler
func (h *Retrieve) Handler() common.HTTPRequestHandler {
	return h.retrieve
}

// version retrieves the content from the DCAS store by hash
func (h *Retrieve) retrieve(rw http.ResponseWriter, req *http.Request) {
	hash := getHash(req)

	maxSize := getMaxSize(req)

	logger.Debugf("[%s:%s:%s] Retrieving resp for hash [%s] with max-size %d", h.channelID, h.ChaincodeName, h.Collection, hash, maxSize)

	rrw := newRetrieveWriter(rw)

	content, err := h.doRetrieve(hash, maxSize)
	if err != nil {
		rrw.WriteError(err)
		return
	}

	logger.Debugf("[%s:%s:%s] ... retrieved content for hash [%s]: Content: %s", h.channelID, h.ChaincodeName, h.Collection, hash, content)

	rrw.Write(content)
}

func (h *Retrieve) doRetrieve(hash string, maxSize int) ([]byte, error) {
	if hash == "" {
		return nil, newRetrieveError(http.StatusBadRequest, CodeInvalidHash)
	}

	if maxSize == 0 {
		return nil, newRetrieveError(http.StatusBadRequest, CodeMaxSizeNotSpecified)
	}

	content, err := h.retrieveContent(hash)
	if err != nil {
		return nil, err
	}

	if maxSize > 0 && len(content) > maxSize {
		return nil, newRetrieveError(http.StatusBadRequest, CodeMaxSizeExceeded)
	}

	return content, nil
}

func (h *Retrieve) retrieveContent(hash string) ([]byte, error) {
	dcasClient, err := h.dcasProvider.GetDCASClient(h.channelID, h.ChaincodeName, h.Collection)
	if err != nil {
		logger.Errorf("[%s:%s:%s] Could not get DCAS client: %s", h.channelID, h.ChaincodeName, h.Collection, err)

		return nil, newRetrieveError(http.StatusInternalServerError, CodeCasNotReachable)
	}

	content := bytes.NewBuffer(nil)

	err = dcasClient.Get(hash, content)
	if err != nil {
		if strings.Contains(err.Error(), encodingErrMsg) {
			logger.Debugf("[%s:%s:%s] Error retrieving DCAS document for hash [%s]: %s", h.channelID, h.ChaincodeName, h.Collection, hash, err)

			return nil, newRetrieveError(http.StatusBadRequest, CodeInvalidHash)
		}

		logger.Errorf("[%s:%s:%s] Error retrieving DCAS document for hash [%s]: %s", h.channelID, h.ChaincodeName, h.Collection, hash, err)

		return nil, newRetrieveError(http.StatusInternalServerError, CodeCasNotReachable)
	}

	if len(content.Bytes()) == 0 {
		logger.Debugf("[%s:%s:%s] Content not found in DCAS for hash [%s]", h.channelID, h.ChaincodeName, h.Collection, hash)

		return nil, newRetrieveError(http.StatusNotFound, CodeNotFound)
	}

	return content.Bytes(), nil
}

var getHash = func(req *http.Request) string {
	return mux.Vars(req)[hashParam]
}

var getMaxSize = func(req *http.Request) int {
	params := getParams(req)

	values := params[maxSizeParam]
	if len(values) > 0 && values[0] != "" {
		return maxSizeFromString(values[0])
	}

	return 0
}

func maxSizeFromString(str string) int {
	size, err := strconv.Atoi(str)
	if err != nil {
		logger.Debugf("Invalid value for parameter [max-size]: %s", err)

		return 0
	}

	return size
}

type retrieveWriter struct {
	*httpserver.ResponseWriter
}

func newRetrieveWriter(rw http.ResponseWriter) *retrieveWriter {
	return &retrieveWriter{
		ResponseWriter: httpserver.NewResponseWriter(rw),
	}
}

func (rw *retrieveWriter) Write(content []byte) {
	rw.ResponseWriter.Write(http.StatusOK, content, httpserver.ContentTypeBinary)
}

func (rw *retrieveWriter) WriteError(err error) {
	readErr, ok := err.(*retrieveError)
	if ok {
		rw.ResponseWriter.Write(readErr.Status(), []byte(readErr.ResultCode()), httpserver.ContentTypeText)
		return
	}

	rw.ResponseWriter.WriteError(err)
}

var getParams = func(req *http.Request) map[string][]string {
	return req.URL.Query()
}
