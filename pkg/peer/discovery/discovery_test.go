/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package discovery

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/hyperledger/fabric-protos-go/gossip"
	"github.com/stretchr/testify/require"

	"github.com/trustbloc/fabric-peer-ext/pkg/common/requestmgr"
	extmocks "github.com/trustbloc/fabric-peer-ext/pkg/mocks"
	"github.com/trustbloc/sidetree-fabric/pkg/peer/discovery/mocks"
)

//go:generate counterfeiter -o ./mocks/appdatahandler.gen.go --fake-name AppDataHandlerRegistry . appDataHandlerRegistry
//go:generate counterfeiter -o ./mocks/discoveryconfig.gen.go --fake-name DiscoveryConfig . discoveryConfig

const (
	channel1 = "channel1"

	org1    = "Org1MSP"
	peer1_1 = "peer1.org1:48326"
	peer1_2 = "peer2.org1:48326"

	org2    = "Org2MSP"
	peer2_1 = "peer1.org2:48326"
	peer2_2 = "peer2.org2:48326"

	domain1 = "org1.com"

	service1 = "sidetree"
	path1_1  = "/identifiers"
	path1_2  = "/operations"

	service2 = "cas"

	rootEndpoint1_p1_1 = "https://peer1.org1.com/sidetree/v1"
	rootEndpoint1_p1_2 = "https://peer2.org1.com/sidetree/v1"
	rootEndpoint2_p1_1 = "https://peer1.org1.com/sidetree/v1/cas"
	rootEndpoint2_p1_2 = "https://peer2.org1.com/sidetree/v1/cas"

	rootEndpoint1_p2_1 = "https://peer1.org2.com/sidetree/v1"
	rootEndpoint1_p2_2 = "https://peer2.org2.com/sidetree/v1"
	rootEndpoint2_p2_1 = "https://peer1.org2.com/sidetree/v1/cas"
	rootEndpoint2_p2_2 = "https://peer2.org2.com/sidetree/v1/cas"

	v1 = "v1"
	v2 = "v2"
)

var (
	pkiid1_1 = []byte("pkiid1_1")
	pkiid1_2 = []byte("pkiid1_2")
	pkiid2_1 = []byte("pkiid2_1")
	pkiid2_2 = []byte("pkiid2_2")

	s1_p1_1 = NewService(service1, v1, domain1, rootEndpoint1_p1_1,
		NewEndpoint(path1_1, http.MethodGet),
		NewEndpoint(path1_2, http.MethodPost),
	)

	s1_p1_2 = NewService(service1, v1, domain1, rootEndpoint1_p1_2,
		NewEndpoint(path1_1, http.MethodGet),
	)

	s2_p1_1 = NewService(service2, v1, domain1, rootEndpoint2_p1_1,
		NewEndpoint("", http.MethodGet),
		NewEndpoint("", http.MethodPost),
	)

	s2_p1_2 = NewService(service2, v1, domain1, rootEndpoint2_p1_2,
		NewEndpoint("", http.MethodGet),
		NewEndpoint("", http.MethodPost),
	)

	s1_p2_1 = NewService(service1, v1, domain1, rootEndpoint1_p2_1,
		NewEndpoint(path1_1, http.MethodGet),
		NewEndpoint(path1_2, http.MethodPost),
	)

	s1_p2_2 = NewService(service1, v1, domain1, rootEndpoint1_p2_2,
		NewEndpoint(path1_1, http.MethodGet),
	)

	s2_p2_1 = NewService(service2, v1, domain1, rootEndpoint2_p2_1,
		NewEndpoint("", http.MethodGet),
		NewEndpoint("", http.MethodPost),
	)

	s2_p2_2 = NewService(service2, v1, domain1, rootEndpoint2_p2_2,
		NewEndpoint("", http.MethodGet),
		NewEndpoint("", http.MethodPost),
	)
)

func TestNew(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		r := &mocks.AppDataHandlerRegistry{}
		d := New(r, &extmocks.GossipProvider{}, &mockPeerConfig{address: peer1_1}, &mocks.DiscoveryConfig{})
		require.NotNil(t, d)
	})

	t.Run("Register error -> panic", func(t *testing.T) {
		errExpected := errors.New("registry error")
		r := &mocks.AppDataHandlerRegistry{}
		r.RegisterReturns(errExpected)

		require.PanicsWithValue(t, errExpected, func() {
			New(r, &extmocks.GossipProvider{}, &mockPeerConfig{address: peer1_1}, &mocks.DiscoveryConfig{})
		})
	})
}

func TestDiscovery_ServicesForChannel(t *testing.T) {
	r := &mocks.AppDataHandlerRegistry{}

	gossip := extmocks.NewMockGossipAdapter().
		Self(org1, extmocks.NewMember(peer1_1, pkiid1_1)).
		Member(org1, extmocks.NewMember(peer1_2, pkiid1_2)).
		Member(org2, extmocks.NewMember(peer2_1, pkiid2_1)).
		Member(org2, extmocks.NewMember(peer2_2, pkiid2_2))

	gp := &extmocks.GossipProvider{}
	gp.GetGossipServiceReturns(gossip)

	peerCfg := &mockPeerConfig{address: peer1_1}

	discoveryCfg := &mocks.DiscoveryConfig{}
	discoveryCfg.DiscoveryGossipTimeoutReturns(1000 * time.Millisecond)
	discoveryCfg.DiscoveryGossipMaxAttemptsReturns(1)

	d := New(r, gp, peerCfg, discoveryCfg)
	require.NotNil(t, d)

	services1_1 := []Service{s1_p1_1, s2_p1_1}
	services1_2 := []Service{s1_p1_2, s2_p1_2}
	services2_1 := []Service{s1_p2_1, s2_p2_1}
	services2_2 := []Service{s1_p2_2, s2_p2_2}

	d.UpdateLocalServicesForChannel(channel1, services1_1)

	reqMgr := requestmgr.Get(channel1)
	req := reqMgr.NewRequest()
	logger.Infof("Will respond to requests > %d", req.ID())

	servicesMap1_2 := make(servicesByPeer)
	servicesMap1_2[peer1_2] = services1_2
	servicesMap1_2[peer2_1] = services2_1

	servicesMap2_1 := make(servicesByPeer)
	servicesMap2_1[peer2_1] = services2_1
	servicesMap2_1[peer2_2] = services2_2

	servicesMap2_2 := make(servicesByPeer)
	servicesMap2_2[peer2_2] = services2_2

	resp1_2, err := json.Marshal(servicesMap1_2)
	require.NoError(t, err)
	resp2_1, err := json.Marshal(servicesMap2_1)
	require.NoError(t, err)
	resp2_2, err := json.Marshal(servicesMap2_2)
	require.NoError(t, err)

	go func() {
		time.Sleep(50 * time.Millisecond)
		requestmgr.Get(channel1).Respond(req.ID(), nil)
		requestmgr.Get(channel1).Respond(req.ID()+1, &requestmgr.Response{Endpoint: peer1_2, MSPID: org1, Data: resp1_2})
		requestmgr.Get(channel1).Respond(req.ID()+2, &requestmgr.Response{Endpoint: peer2_1, MSPID: org2, Data: resp2_1})
		requestmgr.Get(channel1).Respond(req.ID()+3, &requestmgr.Response{Endpoint: peer2_2, MSPID: org2, Data: resp2_2})
	}()

	ch1Services := d.ServicesForChannel(channel1)
	require.Len(t, ch1Services, 8)
	require.True(t, containsAll(ch1Services, s1_p1_1, s1_p1_2, s1_p2_1, s1_p2_2, s2_p1_1, s2_p1_2, s2_p2_1, s2_p2_2))
}

func TestDiscovery_HandleDataRequest(t *testing.T) {
	d := New(&mocks.AppDataHandlerRegistry{}, &extmocks.GossipProvider{}, &mockPeerConfig{address: peer1_1}, &mocks.DiscoveryConfig{})
	require.NotNil(t, d)

	t.Run("Success", func(t *testing.T) {
		reqBytes, err := json.Marshal([]string{peer1_1, peer1_2, peer2_1})
		require.NoError(t, err)

		req := &gossip.AppDataRequest{
			Nonce:    10000,
			DataType: handlerDataType,
			Request:  reqBytes,
		}

		responder := &mockResponder{}

		d.handleDiscoveryRequest(channel1, req, responder)
		require.NotEmpty(t, responder.data)
	})

	t.Run("Marshal error", func(t *testing.T) {
		errExpected := fmt.Errorf("injected marshal error")

		d.marshal = func(interface{}) ([]byte, error) { return nil, errExpected }

		reqBytes, err := json.Marshal([]string{peer1_1, peer1_2, peer2_1})
		require.NoError(t, err)

		req := &gossip.AppDataRequest{
			Nonce:    10000,
			DataType: handlerDataType,
			Request:  reqBytes,
		}

		responder := &mockResponder{}

		d.handleDiscoveryRequest(channel1, req, responder)
		require.Empty(t, responder.data)
	})

	t.Run("Unmarshal error", func(t *testing.T) {
		req := &gossip.AppDataRequest{
			Nonce:    10000,
			DataType: handlerDataType,
			Request:  []byte("{"),
		}

		responder := &mockResponder{}

		d.handleDiscoveryRequest(channel1, req, responder)
		require.Empty(t, responder.data)
	})
}

type mockPeerConfig struct {
	address string
}

func (p *mockPeerConfig) PeerAddress() string {
	return p.address
}

func containsAll(services []Service, svcs ...Service) bool {
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

type mockResponder struct {
	data []byte
}

func (m *mockResponder) Respond(data []byte) {
	m.data = data
}
