/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package sidetreesvc

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/trustbloc/sidetree-fabric/pkg/peer/mocks"
)

const (
	service1  = "did:sidetree"
	v1        = "0.0.1"
	basePath1 = "/sidetree/0.0.1"
	ep1       = "/identifiers"
	ep2       = "/operations"
)

func TestService(t *testing.T) {
	h1 := &mocks.HTTPHandler{}
	h1.MethodReturns(http.MethodGet)

	h2 := &mocks.HTTPHandler{}
	h2.MethodReturns(http.MethodPost)

	s := newService(service1, v1, basePath1,
		newEndpoint(ep1, h1),
		newEndpoint(ep2, h2),
	)

	require.NotNil(t, s)
	require.Equal(t, service1, s.name)
	require.Equal(t, basePath1, s.basePath)
	require.Equal(t, v1, s.apiVersion)
	require.Len(t, s.endpoints, 2)
	require.Equal(t, ep1, s.endpoints[0].name)
	require.Equal(t, ep2, s.endpoints[1].name)

	h := s.endpoints.Handlers()
	require.Len(t, h, 2)
	require.Equal(t, http.MethodGet, h[0].Method())
	require.Equal(t, http.MethodPost, h[1].Method())
}
