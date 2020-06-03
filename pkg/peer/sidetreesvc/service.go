/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package sidetreesvc

import (
	"github.com/trustbloc/sidetree-core-go/pkg/restapi/common"
)

type service struct {
	name       string
	apiVersion string
	basePath   string
	endpoints  endpoints
}

func newService(name, apiVersion, basePath string, endpoints ...*endpoint) *service {
	return &service{
		name:       name,
		apiVersion: apiVersion,
		basePath:   basePath,
		endpoints:  endpoints,
	}
}

type endpoint struct {
	name    string
	method  string
	handler common.HTTPHandler
}

func newEndpoint(name string, handler common.HTTPHandler) *endpoint {
	return &endpoint{
		name:    name,
		method:  handler.Method(),
		handler: handler,
	}
}

type endpoints []*endpoint

func (s endpoints) Handlers() []common.HTTPHandler {
	handlers := make([]common.HTTPHandler, len(s))

	for i, api := range s {
		handlers[i] = api.handler
	}

	return handlers
}
