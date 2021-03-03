/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package discovery

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewService(t *testing.T) {
	const (
		domain1         = "org1.com"
		service1        = "service1"
		path1_1         = "/identifiers"
		path1_2         = "/operations"
		rootEndpoint1_1 = "https://peer1.org1.com/sidetree/v1"
		v1              = "0.0.1"
	)

	s := NewService(service1, v1, domain1, rootEndpoint1_1,
		NewEndpoint(path1_1, http.MethodGet),
		NewEndpoint(path1_2, http.MethodPost),
	)

	require.NotNil(t, s)
	require.Equal(t, service1, s.Service)
	require.Equal(t, v1, s.APIVersion)
	require.Equal(t, domain1, s.Domain)
	require.Equal(t, rootEndpoint1_1, s.RootEndpoint)
	require.Len(t, s.Endpoints, 2)
	require.Equal(t, path1_1, s.Endpoints[0].Path)
	require.Equal(t, http.MethodGet, s.Endpoints[0].Method)
	require.Equal(t, path1_2, s.Endpoints[1].Path)
	require.Equal(t, http.MethodPost, s.Endpoints[1].Method)
}
