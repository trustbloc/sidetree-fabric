/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package lease

import (
	"testing"

	gcommon "github.com/hyperledger/fabric/gossip/common"
	"github.com/stretchr/testify/require"
	"github.com/trustbloc/fabric-peer-ext/pkg/mocks"
	"github.com/trustbloc/fabric-peer-ext/pkg/roles"
	"github.com/trustbloc/sidetree-fabric/pkg/role"
)

const (
	channel1 = "channel1"

	org1MSPID = "Org1MSP"
	org2MSPID = "Org2MSP"

	p1Org1Endpoint = "p1.org1.com"
	p2Org1Endpoint = "p2.org1.com"
	p3Org1Endpoint = "p3.org1.com"

	p1Org2Endpoint = "p1.org2.com"
)

var (
	p1Org1PKIID = gcommon.PKIidType("pkiid_P1O1")
	p2Org1PKIID = gcommon.PKIidType("pkiid_P2O1")
	p3Org1PKIID = gcommon.PKIidType("pkiid_P3O1")

	p1Org2PKIID = gcommon.PKIidType("pkiid_P1O2")

	p1Org1 = mocks.NewMember(p1Org1Endpoint, p1Org1PKIID)
	p2Org1 = mocks.NewMember(p2Org1Endpoint, p2Org1PKIID, role.Observer)
	p3Org1 = mocks.NewMember(p3Org1Endpoint, p3Org1PKIID, role.ObserverStandby)

	p1Org2 = mocks.NewMember(p1Org2Endpoint, p1Org2PKIID, role.Observer)

	// Ensure roles are initialized
	_ = roles.GetRoles()
)

func TestLease(t *testing.T) {
	roles.SetRoles(map[roles.Role]struct{}{roles.CommitterRole: {}, role.ObserverStandby: {}})
	defer roles.SetRoles(nil)

	require.True(t, roles.IsClustered())

	t.Run("GetLease", func(t *testing.T) {
		p := NewProvider(channel1, mocks.NewMockGossipAdapter())

		l := p.GetLease(p1Org1Endpoint)
		require.NotNil(t, l)
		require.Equal(t, p1Org1Endpoint, l.Owner())
	})

	t.Run("Choose active observer", func(t *testing.T) {
		gossip := mocks.NewMockGossipAdapter()
		gossip.Self(org1MSPID, p1Org1).
			Member(org1MSPID, p2Org1).
			Member(org1MSPID, p3Org1).
			Member(org2MSPID, p1Org2)
		p := NewProvider(channel1, gossip)

		l := p.CreateLease(1001)
		require.NotNil(t, l)
		require.Equal(t, p2Org1Endpoint, l.Owner())
	})

	t.Run("Active observer down", func(t *testing.T) {
		gossip := mocks.NewMockGossipAdapter()
		gossip.Self(org1MSPID, p1Org1).
			Member(org1MSPID, p3Org1)
		p := NewProvider(channel1, gossip)

		l := p.CreateLease(1000)
		require.NotNil(t, l)
		require.Equal(t, p1Org1Endpoint, l.Owner())
		require.True(t, l.IsLocalPeerOwner())
		require.True(t, l.IsValid())

		l = p.CreateLease(1001)
		require.NotNil(t, l)
		require.Equal(t, p3Org1Endpoint, l.Owner())
		require.False(t, l.IsLocalPeerOwner())
		require.True(t, l.IsValid())
	})
}

func TestLease_NotClustered(t *testing.T) {
	require.False(t, roles.IsClustered())

	gossip := mocks.NewMockGossipAdapter()
	gossip.Self(org1MSPID, p1Org1).
		Member(org1MSPID, p2Org1).
		Member(org1MSPID, p3Org1)
	p := NewProvider(channel1, gossip)

	l := p.CreateLease(1001)
	require.NotNil(t, l)
	require.Equal(t, p1Org1Endpoint, l.Owner())
	require.True(t, l.IsValid())
}
