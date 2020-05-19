/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package lease

import (
	"github.com/hyperledger/fabric/common/flogging"
	gapi "github.com/hyperledger/fabric/gossip/api"
	gcommon "github.com/hyperledger/fabric/gossip/common"
	gdiscovery "github.com/hyperledger/fabric/gossip/discovery"
	"github.com/trustbloc/fabric-peer-ext/pkg/common/discovery"
	"github.com/trustbloc/fabric-peer-ext/pkg/roles"

	"github.com/trustbloc/sidetree-fabric/pkg/role"
)

var logger = flogging.MustGetLogger("sidetree_observer")

type isValidFunc func(l *Lease) bool

// Lease specifies the peer that is the active Observer in the cluster
type Lease struct {
	self    string
	owner   string
	isValid isValidFunc
}

// Owner returns the owner of the lease
func (l *Lease) Owner() string {
	return l.owner
}

// IsLocalPeerOwner returns true if the local peer is the owner of this lease
func (l *Lease) IsLocalPeerOwner() bool {
	return l.owner == l.self
}

// IsValid returns true if this lease is still valid. In order to be valid, the owner needs to
// be 'alive' (according to Gossip discovery) and, if the owner is a 'standby' observer, there are no
// 'active' observers that are alive.
func (l *Lease) IsValid() bool {
	return l.isValid(l)
}

type gossipService interface {
	PeersOfChannel(id gcommon.ChannelID) []gdiscovery.NetworkMember
	SelfMembershipInfo() gdiscovery.NetworkMember
	IdentityInfo() gapi.PeerIdentitySet
}

// Provider is a lease provider
type Provider struct {
	*discovery.Discovery
	isValid     isValidFunc
	createLease func(blockNum uint64) *Lease
	self        *discovery.Member
}

// NewProvider returns a new lease provider
func NewProvider(channelID string, gossip gossipService) *Provider {
	p := &Provider{
		Discovery: discovery.New(channelID, gossip),
	}

	if roles.IsClustered() {
		p.isValid = p.isValidClustered
		p.createLease = p.createLeaseClustered
	} else {
		p.isValid = p.isValidNoCluster
		p.createLease = p.createLeaseNoCluster
	}

	p.self = p.Self()

	return p
}

// GetLease returns a lease with the specified owner
func (p *Provider) GetLease(owner string) *Lease {
	return &Lease{
		self:    p.self.Endpoint,
		owner:   owner,
		isValid: p.isValid,
	}
}

// CreateLease creates a lease for the given block number. If the peer is not operating in clustered mode then
// the local peer is always the lease owner. In clustered-mode, the owner of the lease is chosen from all peers
// with either the sidetree-observer or sidetree-observer-standby role, where the sidetree-observer role is given priority.
// If there are multiple peers with the same role then the peer is deterministically chosen from these peers based
// on the given block number so that each peer in the cluster resolves to the same peer.
func (p *Provider) CreateLease(blockNum uint64) *Lease {
	return p.createLease(blockNum)
}

func (p *Provider) createLeaseClustered(blockNum uint64) *Lease {
	peers := p.observers().Sort()

	logger.Debugf("[%s] All observers: %s", p.ChannelID(), peers)

	owner := peers[blockNum%uint64(len(peers))]

	logger.Debugf("[%s] Chosen lease owner for block %d: %s", p.ChannelID(), blockNum, owner)

	return &Lease{
		self:    p.self.Endpoint,
		owner:   owner.Endpoint,
		isValid: p.isValid,
	}
}

func (p *Provider) createLeaseNoCluster(blockNum uint64) *Lease {
	return &Lease{
		self:    p.self.Endpoint,
		owner:   p.self.Endpoint,
		isValid: p.isValid,
	}
}

// isValidClustered returns true if the given lease is still valid. In order to be valid, the owner needs to
// be 'alive' (according to Gossip discovery) and, if the owner is a 'standby' observer, there are no
// 'active' observers that are alive.
func (p *Provider) isValidClustered(l *Lease) bool {
	return p.observers().Contains(&discovery.Member{
		NetworkMember: gdiscovery.NetworkMember{
			Endpoint: l.Owner(),
		},
	})
}

// isValidNoCluster returns true if the least is held by the local peer (since in non-clustered
// mode the owner is always the local peer)
func (p *Provider) isValidNoCluster(l *Lease) bool {
	return l.owner == l.self
}

func (p *Provider) observers() discovery.PeerGroup {
	observers := p.peersInRole(role.Observer)
	if observers.Len() == 0 {
		observers = p.peersInRole(role.ObserverStandby)

		logger.Infof("[%s] Using standby observers since all of the active observers are down: %s", p.ChannelID(), observers)
	} else {
		logger.Debugf("[%s] Got active observers: %s", p.ChannelID(), observers)
	}

	return observers
}

func (p *Provider) peersInRole(role roles.Role) discovery.PeerGroup {
	return p.Discovery.GetMembers(func(peer *discovery.Member) bool {
		if peer.MSPID != p.self.MSPID {
			return false
		}

		if !peer.HasRole(role) {
			return false
		}

		return true
	})
}
