/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package discovery

/*
A service is described by a name, API version, root-endpoint, domain name, and one or more
endpoints off the root-endpoint. Each endpoint contains an HTTP method and optional path.
Following are some JSON examples:

[
  {
    "service": "did:sidetree",
    "apiVersion": "0.0.1",
    "domain": "org1.com",
    "rootEndpoint": "https://peer0.org1.com:48326/sidetree/0.0.1",
    "endpoints": [
      {
        "path": "/identifiers",
        "method": "GET"
      },
      {
        "path": "/operations",
        "method": "POST"
      },
      {
	    "path": "/version",
	    "method": "GET"
      }
    ]
  },
  {
    "service": "cas",
    "apiVersion": "0.0.1",
    "domain": "org2.com",
    "rootEndpoint": "https://peer0.org2.com:48326/sidetree/0.0.1/cas",
    "endpoints": [
      {
	    "method": "GET"
      },
      {
	    "method": "POST"
      },
      {
	    "path": "/version",
	    "method": "GET"
      }
    ]
  }
]
*/

// Service contains information about one or more endpoints for a service
type Service struct {
	Service      string     `json:"service"`
	APIVersion   string     `json:"apiVersion,omitempty"`
	Domain       string     `json:"domain"`
	RootEndpoint string     `json:"rootEndpoint"`
	Endpoints    []Endpoint `json:"endpoints"`
}

// NewService returns a new Service
func NewService(name, apiVersion, domain, rootEndpoint string, endpoints ...Endpoint) Service {
	return Service{
		Service:      name,
		APIVersion:   apiVersion,
		Domain:       domain,
		RootEndpoint: rootEndpoint,
		Endpoints:    endpoints,
	}
}

// Endpoint contains the path and HTTP method of the REST endpoint
type Endpoint struct {
	Path   string `json:"path,omitempty"`
	Method string `json:"method"`
}

// NewEndpoint returns a new Endpoint
func NewEndpoint(path, method string) Endpoint {
	return Endpoint{
		Path:   path,
		Method: method,
	}
}
