/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package discoveryhandler

import (
	"github.com/trustbloc/sidetree-fabric/pkg/peer/discovery"
)

const (
	serviceFilterParam    = "service"
	domainFilterParam     = "domain"
	apiVersionFilterParam = "api-version"
	pathFilterParam       = "path"
	methodFilterParam     = "method"
)

type getFilterFunc func(...string) Filter

var filterByParam = map[string]getFilterFunc{
	serviceFilterParam:    ByService,
	domainFilterParam:     ByDomain,
	apiVersionFilterParam: ByAPIVersion,
	pathFilterParam:       ByPath,
	methodFilterParam:     ByMethod,
}

// Services is a slice of discovery services
type Services []discovery.Service

// FilterByParams removes services from the slice according to the given set of filter params
func (s Services) FilterByParams(params map[string][]string) Services {
	var filters Filters
	for param, getFilter := range filterByParam {
		values := params[param]
		if len(values) > 0 {
			filters = append(filters, getFilter(values...))
		}
	}

	return s.filter(filters...)
}

func (s Services) filter(f ...Filter) Services {
	var svcs Services

	for _, s := range s {
		if Filters(f).Accept(s) {
			svcs = append(svcs, s)
		}
	}

	return svcs
}

// Filter is a discovery service filter
type Filter func(s discovery.Service) bool

// Filters is a slice of filters
type Filters []Filter

// Accept returns true if the given service should be included
func (fltrs Filters) Accept(s discovery.Service) bool {
	for _, f := range fltrs {
		if !f(s) {
			return false
		}
	}

	return true
}

// ByService is a service filter that returns true only if the service has the given name
func ByService(values ...string) Filter {
	return func(s discovery.Service) bool {
		return contains(values, s.Service)
	}
}

// ByDomain is a service filter that returns true only if the service has the given domain
func ByDomain(values ...string) Filter {
	return func(s discovery.Service) bool {
		return contains(values, s.Domain)
	}
}

// ByAPIVersion is a service filter that returns true only if the service has the given API version
func ByAPIVersion(values ...string) Filter {
	return func(s discovery.Service) bool {
		return contains(values, s.APIVersion)
	}
}

// ByPath is a service filter that returns true only if the service has the given path
func ByPath(values ...string) Filter {
	return func(s discovery.Service) bool {
		for _, e := range s.Endpoints {
			if contains(values, e.Path) {
				return true
			}
		}

		return false
	}
}

// ByMethod is a service filter that returns true only if the service has the given HTTP method
func ByMethod(values ...string) Filter {
	return func(s discovery.Service) bool {
		for _, e := range s.Endpoints {
			if contains(values, e.Method) {
				return true
			}
		}

		return false
	}
}

func contains(values []string, value string) bool {
	for _, v := range values {
		if v == value {
			return true
		}
	}

	return false
}
