/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package discoveryhandler

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/trustbloc/sidetree-fabric/pkg/peer/discovery"
)

const (
	domain1 = "org1.com"
	domain2 = "org2.com"

	service1           = "service1"
	path1_1            = "/identifiers"
	path1_2            = "/operations"
	path1_3            = "/version"
	rootEndpoint1_1    = "https://peer1.org1.com/sidetree/0.0.1"
	rootEndpoint1_1_v2 = "https://peer1.org1.com/sidetree/0.0.2"
	rootEndpoint1_2    = "https://peer1.org2.com/sidetree/0.0.1"

	service2           = "service2"
	rootEndpoint2_1    = "https://peer1.org1.com/sidetree/0.0.1/cas"
	rootEndpoint2_1_v2 = "https://peer1.org1.com/sidetree/0.0.2/cas"
	rootEndpoint2_2    = "https://peer2.org2.com/sidetree/0.0.1/cas"
	rootEndpoint2_2_v2 = "https://peer2.org2.com/sidetree/0.0.2/cas"

	v1 = "0.0.1"
	v2 = "0.0.2"
)

var (
	s1_1 = discovery.NewService(service1, v1, domain1, rootEndpoint1_1,
		discovery.NewEndpoint(path1_1, http.MethodGet),
		discovery.NewEndpoint(path1_2, http.MethodPost),
		discovery.NewEndpoint(path1_3, http.MethodGet),
	)

	s1_1_v2 = discovery.NewService(service1, v2, domain1, rootEndpoint1_1_v2,
		discovery.NewEndpoint(path1_1, http.MethodGet),
		discovery.NewEndpoint(path1_2, http.MethodPost),
	)

	s1_2 = discovery.NewService(service1, v1, domain2, rootEndpoint1_2,
		discovery.NewEndpoint(path1_1, http.MethodGet),
	)

	s2_1 = discovery.NewService(service2, v1, domain1, rootEndpoint2_1,
		discovery.NewEndpoint("", http.MethodGet),
		discovery.NewEndpoint("", http.MethodPost),
	)

	s2_1_v2 = discovery.NewService(service2, v2, domain2, rootEndpoint2_1_v2,
		discovery.NewEndpoint("", http.MethodPost),
	)

	s2_2 = discovery.NewService(service2, v1, domain2, rootEndpoint2_2,
		discovery.NewEndpoint("", http.MethodGet),
		discovery.NewEndpoint("", http.MethodPost),
	)

	s2_2_v2 = discovery.NewService(service2, v2, domain2, rootEndpoint2_2_v2,
		discovery.NewEndpoint("", http.MethodGet),
		discovery.NewEndpoint("", http.MethodPost),
	)
)

func TestServices_Filter(t *testing.T) {
	services := Services{s1_1, s1_2, s2_1, s2_2}

	t.Run("Service", func(t *testing.T) {
		service2Services := services.filter(ByService(service2))

		require.True(t, containsAll(service2Services, s2_1, s2_1_v2, s2_2, s2_2_v2))
		require.False(t, containsAny(service2Services, s1_1, s1_1_v2, s1_2))
	})

	t.Run("Domain", func(t *testing.T) {
		domain1Services := services.filter(ByDomain(domain1))

		require.True(t, containsAll(domain1Services, s1_1, s1_1_v2, s2_1, s2_1_v2))
		require.False(t, containsAny(domain1Services, s1_2, s2_2, s2_2_v2))
	})

	t.Run("ApiVersion", func(t *testing.T) {
		v1Services := services.filter(ByAPIVersion(v1))

		require.True(t, containsAll(v1Services, s1_1, s1_2, s2_1, s2_2))
		require.False(t, containsAny(v1Services, s1_1_v2, s2_1_v2, s2_2_v2))
	})

	t.Run("Method", func(t *testing.T) {
		postServices := services.filter(ByMethod(http.MethodPost))

		require.True(t, containsAll(postServices, s1_1, s1_1_v2, s2_1, s2_1_v2, s2_2, s2_2_v2))
		require.False(t, containsAny(postServices, s1_2))
	})

	t.Run("Path", func(t *testing.T) {
		postServices := services.filter(ByPath(path1_1, path1_3))

		require.True(t, containsAll(postServices, s1_1, s1_1_v2, s1_2))
		require.False(t, containsAny(postServices, s2_1, s2_1_v2, s2_2, s2_2_v2))
	})

	t.Run("Service, Method, ApiVersion", func(t *testing.T) {
		filteredServices := services.filter(ByService(service2), ByAPIVersion(v2), ByMethod(http.MethodPost))

		require.True(t, containsAll(filteredServices, s2_2_v2))
		require.False(t, containsAny(filteredServices, s1_1, s1_1_v2, s1_2, s2_1, s2_2, s2_1_v2))
	})
}

func TestServices_FilterByParams(t *testing.T) {
	services := Services{s1_1, s1_2, s2_1, s2_2}

	t.Run("Service", func(t *testing.T) {
		filteredServices := services.FilterByParams(map[string][]string{serviceFilterParam: {service2}})
		require.True(t, containsAll(filteredServices, s2_1, s2_1_v2, s2_2, s2_2_v2))
		require.False(t, containsAny(filteredServices, s1_1, s1_1_v2, s1_2))

		allServices := services.FilterByParams(map[string][]string{serviceFilterParam: {service1, service2}})
		require.True(t, containsAll(allServices, s1_1, s1_1_v2, s1_2, s2_1, s2_1_v2, s2_2, s2_2_v2))
	})

	t.Run("Service, Method, ApiVersion", func(t *testing.T) {
		params := map[string][]string{
			serviceFilterParam:    {service2},
			apiVersionFilterParam: {v2},
			methodFilterParam:     {http.MethodPost},
		}

		filteredServices := services.FilterByParams(params)

		require.True(t, containsAll(filteredServices, s2_2_v2))
		require.False(t, containsAny(filteredServices, s1_1, s1_1_v2, s1_2, s2_1, s2_2, s2_1_v2))
	})
}

func containsAll(services []discovery.Service, svcs ...discovery.Service) bool {
	for _, s := range services {
		contains := false

		for _, svc := range svcs {
			if reflect.DeepEqual(s, svc) {
				contains = true
				break
			}
		}

		if !contains {
			return false
		}
	}

	return true
}

func containsAny(services []discovery.Service, svcs ...discovery.Service) bool {
	for _, s := range services {
		for _, svc := range svcs {
			if reflect.DeepEqual(s, svc) {
				return true
			}
		}
	}

	return false
}
