/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package discovery

import (
	"context"
	"encoding/json"
	"time"

	"github.com/bluele/gcache"
	gproto "github.com/hyperledger/fabric-protos-go/gossip"
	"github.com/hyperledger/fabric/common/flogging"
	gossipapi "github.com/hyperledger/fabric/extensions/gossip/api"
	"github.com/hyperledger/fabric/gossip/api"
	"github.com/hyperledger/fabric/gossip/comm"
	gcommon "github.com/hyperledger/fabric/gossip/common"
	"github.com/hyperledger/fabric/gossip/discovery"
	"github.com/pkg/errors"
	extcommon "github.com/trustbloc/fabric-peer-ext/pkg/common"
	extdiscovery "github.com/trustbloc/fabric-peer-ext/pkg/common/discovery"
	"github.com/trustbloc/fabric-peer-ext/pkg/gossip/appdata"
)

var logger = flogging.MustGetLogger("sidetree_peer")

const (
	handlerDataType = "ENDPOINT_DISCOVERY"

	defaultCacheExpirationTime = 5 * time.Minute
	defaultGossipTimeout       = 5 * time.Second
	defaultGossipMaxAttempts   = 5
	defaultGossipMaxPeers      = 3
)

type appDataHandlerRegistry interface {
	Register(dataType string, handler appdata.Handler) error
}

type gossipService interface {
	PeersOfChannel(id gcommon.ChannelID) []discovery.NetworkMember
	SelfMembershipInfo() discovery.NetworkMember
	IdentityInfo() api.PeerIdentitySet
	Send(msg *gproto.GossipMessage, peers ...*comm.RemotePeer)
}

type gossipProvider interface {
	GetGossipService() gossipapi.GossipService
}

type peerConfig interface {
	PeerAddress() string
}

type discoveryConfig interface {
	DiscoveryCacheExpirationTime() time.Duration
	DiscoveryGossipMaxAttempts() int
	DiscoveryGossipTimeout() time.Duration
	DiscoveryGossipMaxPeers() int
}

type retriever interface {
	Retrieve(ctxt context.Context, request *appdata.Request, responseHandler appdata.ResponseHandler, allSet appdata.AllSet) (extcommon.Values, error)
}

type servicesByPeer = map[string][]Service

type config struct {
	gossipTimeout       time.Duration
	gossipMaxAttempts   int
	gossipMaxPeers      int
	cacheExpirationTime time.Duration
}

// Discovery is a service that discovers all service endpoints in a consortium of peers
type Discovery struct {
	config
	peerConfig
	discoveryCache gcache.Cache
}

// New returns a discovery service
func New(handlerRegistry appDataHandlerRegistry, gossipProvider gossipProvider, peerCfg peerConfig, discoveryCfg discoveryConfig) *Discovery {
	cfg := newConfig(discoveryCfg)

	logger.Infof("Creating discovery service with config: %+v", cfg)

	d := &Discovery{
		config:     cfg,
		peerConfig: peerCfg,
		discoveryCache: gcache.New(0).LoaderFunc(func(chID interface{}) (interface{}, error) {
			return newChannelDiscovery(chID.(string), cfg, peerCfg, gossipProvider.GetGossipService()), nil
		}).Build(),
	}

	if err := handlerRegistry.Register(handlerDataType, d.handleDiscoveryRequest); err != nil {
		panic(err)
	}

	return d
}

// ServicesForChannel returns all services available on the given channel
func (d *Discovery) ServicesForChannel(channelID string) []Service {
	return d.forChannel(channelID).services()
}

// UpdateLocalServicesForChannel updates the cache of services for the local peer
func (d *Discovery) UpdateLocalServicesForChannel(channelID string, services []Service) {
	d.forChannel(channelID).updateLocalServices(services)
}

func (d *Discovery) forChannel(channelID string) *channelDiscovery {
	c, err := d.discoveryCache.Get(channelID)
	if err != nil {
		// Should never happen
		panic(err)
	}

	return c.(*channelDiscovery)
}

// handleDiscoveryRequest is a server-side handler that responds to a remote peer with the requested service endpoints
func (d *Discovery) handleDiscoveryRequest(channelID string, req *gproto.AppDataRequest) ([]byte, error) {
	logger.Debugf("[%s] Handling discovery request", channelID)

	var peerEndpoints []string
	err := json.Unmarshal(req.Request, &peerEndpoints)
	if err != nil {
		return nil, errors.WithMessagef(err, "error unmarshalling peer endpoints in request")
	}

	logger.Debugf("[%s] Got discovery request for endpoints: %s", channelID, peerEndpoints)

	response := make(servicesByPeer)

	for _, peerEndpoint := range peerEndpoints {
		response[peerEndpoint] = d.forChannel(channelID).servicesForPeer(peerEndpoint)
	}

	respBytes, err := json.Marshal(response)
	if err != nil {
		return nil, errors.WithMessagef(err, "error marshalling response")
	}

	return respBytes, nil
}

// handleDiscoveryResponse is a client-side handler of a response from a remote peer. It ensures that
// the response from the peer is added to Values in the same order as the given peer endpoints.
func (c *channelDiscovery) handleDiscoveryResponse(peerEndpoints []string) appdata.ResponseHandler {
	return func(response []byte) (extcommon.Values, error) {
		servicesMap := make(servicesByPeer)
		err := json.Unmarshal(response, &servicesMap)
		if err != nil {
			return nil, err
		}

		services := make(extcommon.Values, len(peerEndpoints))
		for i, k := range peerEndpoints {
			servicesForPeer, ok := servicesMap[k]
			if !ok || servicesForPeer == nil {
				services[i] = nil
			} else {
				services[i] = servicesForPeer
			}
		}

		return services, nil
	}
}

type channelDiscovery struct {
	config
	peerConfig
	*extdiscovery.Discovery
	retriever
	channelID     string
	servicesCache gcache.Cache
}

func newChannelDiscovery(channelID string, cfg config, peerCfg peerConfig, gossip gossipService) *channelDiscovery {
	logger.Infof("[%s] Creating discovery service for channel", channelID)

	return &channelDiscovery{
		channelID:     channelID,
		config:        cfg,
		peerConfig:    peerCfg,
		servicesCache: gcache.New(0).Build(),
		retriever:     appdata.NewRetriever(channelID, gossip, cfg.gossipMaxAttempts, cfg.gossipMaxPeers),
		Discovery:     extdiscovery.New(channelID, gossip),
	}
}

func (c *channelDiscovery) updateLocalServices(services []Service) {
	logger.Debugf("[%s] Updating local services: %v", c.channelID, services)

	// No cache expiration for local services
	if err := c.servicesCache.Set(c.PeerAddress(), services); err != nil {
		// Should never happen
		panic(err)
	}
}

func (c *channelDiscovery) updateRemoteServices(endpoint string, services []Service) {
	logger.Debugf("[%s] Updating services for [%s]: %v", c.channelID, endpoint, services)

	if err := c.servicesCache.SetWithExpire(endpoint, services, c.cacheExpirationTime); err != nil {
		// Should never happen
		panic(err)
	}
}

func (c *channelDiscovery) servicesForPeer(peerEndpoint string) []Service {
	services, err := c.servicesCache.Get(peerEndpoint)
	if err != nil {
		if err == gcache.KeyNotFoundError {
			return nil
		}

		// Should never happen
		panic(err)
	}

	return services.([]Service)
}

func (c *channelDiscovery) services() []Service {
	var services []Service
	var missingEndpoints []string

	// Load all cached services of the peers that are alive (including the local peer).
	for _, p := range extdiscovery.PeerGroup(c.GetMembers(func(m *extdiscovery.Member) bool { return true })) {
		if s := c.servicesForPeer(p.Endpoint); len(s) == 0 {
			missingEndpoints = append(missingEndpoints, p.Endpoint)
		} else {
			services = append(services, s...)
		}
	}

	// Retrieve the services from the remote peers that weren't cached and merge with cached ones
	return append(services, c.servicesForPeers(missingEndpoints)...)
}

func (c *channelDiscovery) servicesForPeers(peerEndpoints []string) []Service {
	if len(peerEndpoints) == 0 {
		return nil
	}

	endpointBytes, err := json.Marshal(peerEndpoints)
	if err != nil {
		logger.Errorf("[%s] Error marshalling data: %s", c.channelID, err)
		return nil
	}

	ctxt, cancel := context.WithTimeout(context.Background(), c.gossipTimeout)
	defer cancel()

	values, err := c.Retrieve(ctxt,
		&appdata.Request{
			DataType: handlerDataType,
			Payload:  endpointBytes,
		},
		c.handleDiscoveryResponse(peerEndpoints),
		func(values extcommon.Values) bool {
			return len(values) == len(peerEndpoints) && values.AllSet()
		},
	)
	if err != nil {
		logger.Errorf("[%s] Got error requesting data from remote peers: %s", c.channelID, err)
		return nil
	}

	var services []Service
	for i, v := range values {
		if v != nil {
			peerServices := v.([]Service)
			services = append(services, peerServices...)

			endpoint := peerEndpoints[i]

			logger.Debugf("[%s] Caching services for peer [%s]", c.channelID, endpoint)

			c.updateRemoteServices(endpoint, peerServices)
		}
	}

	return services
}

func newConfig(discoveryCfg discoveryConfig) config {
	var cfg config

	cfg.cacheExpirationTime = discoveryCfg.DiscoveryCacheExpirationTime()
	if cfg.cacheExpirationTime == 0 {
		cfg.cacheExpirationTime = defaultCacheExpirationTime
	}

	cfg.gossipTimeout = discoveryCfg.DiscoveryGossipTimeout()
	if cfg.gossipTimeout == 0 {
		cfg.gossipTimeout = defaultGossipTimeout
	}

	cfg.gossipMaxAttempts = discoveryCfg.DiscoveryGossipMaxAttempts()
	if cfg.gossipMaxAttempts == 0 {
		cfg.gossipMaxAttempts = defaultGossipMaxAttempts
	}

	cfg.gossipMaxPeers = discoveryCfg.DiscoveryGossipMaxPeers()
	if cfg.gossipMaxPeers == 0 {
		cfg.gossipMaxPeers = defaultGossipMaxPeers
	}

	return cfg
}
