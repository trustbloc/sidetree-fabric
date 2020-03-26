/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package httpserver

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gorilla/mux"
	"github.com/hyperledger/fabric/common/util/retry"
	"github.com/pkg/errors"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
	"github.com/trustbloc/sidetree-core-go/pkg/restapi/common"
)

var logger = logrus.New()

// Server implements an HTTP server
type Server struct {
	httpServer *http.Server
	started    uint32
	certFile   string
	keyFile    string
}

// New returns a new HTTP server
func New(url, certFile, keyFile string, handlers ...common.HTTPHandler) *Server {
	router := mux.NewRouter()
	for _, handler := range handlers {
		logger.Infof("Registering handler for [%s]", handler.Path())
		router.HandleFunc(handler.Path(), handler.Handler()).Methods(handler.Method())
	}

	// TODO configure cors
	handler := cors.Default().Handler(router)

	return &Server{
		httpServer: &http.Server{
			Addr:    url,
			Handler: handler,
		},
		certFile: certFile,
		keyFile:  keyFile,
	}
}

// Start starts the HTTP server in a separate Go routine
func (s *Server) Start() error {
	if !atomic.CompareAndSwapUint32(&s.started, 0, 1) {
		return errors.New("server already started")
	}

	go func() {
		logger.Infof("Listening for requests on [%s]", s.httpServer.Addr)

		err := s.startWithRetry()
		if err != nil && err != http.ErrServerClosed {
			panic(fmt.Sprintf("Failed to start server on [%s]: %s", s.httpServer.Addr, err))
		}
		atomic.StoreUint32(&s.started, 0)
		logger.Infof("Server has stopped")
	}()
	return nil
}

// Stop stops the REST service
func (s *Server) Stop(ctx context.Context) error {
	if !atomic.CompareAndSwapUint32(&s.started, 1, 0) {
		return errors.New("Cannot stop HTTP server since it hasn't been started")
	}
	return s.httpServer.Shutdown(ctx)
}

// startWithRetry will retry to start the HTTP server if a 'address in use' error is experienced. This is
// common during a restart of the HTTP server since it takes a while for the OS to release the port.
func (s *Server) startWithRetry() error {
	_, err := retry.Invoke(
		func() (interface{}, error) {
			if s.keyFile != "" && s.certFile != "" {
				return nil, s.httpServer.ListenAndServeTLS(s.certFile, s.keyFile)
			}
			return nil, s.httpServer.ListenAndServe()
		},
		retry.WithMaxAttempts(10),
		retry.WithBeforeRetry(func(err error, attempt int, backoff time.Duration) bool {
			if strings.Contains(err.Error(), "address already in use") {
				logger.Infof("Got error starting HTTP server on attempt %d. Will retry in %s: %s", attempt, backoff, err)
				return true
			}
			return false
		}),
	)
	return err
}
