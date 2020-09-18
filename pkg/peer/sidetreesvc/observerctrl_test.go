/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package sidetreesvc

import (
	"testing"
	"time"

	gossipapi "github.com/hyperledger/fabric/extensions/gossip/api"
	"github.com/stretchr/testify/require"
	extmocks "github.com/trustbloc/fabric-peer-ext/pkg/mocks"
	extroles "github.com/trustbloc/fabric-peer-ext/pkg/roles"
	coremocks "github.com/trustbloc/sidetree-core-go/pkg/mocks"

	"github.com/trustbloc/sidetree-fabric/pkg/config"
	"github.com/trustbloc/sidetree-fabric/pkg/observer"
	peermocks "github.com/trustbloc/sidetree-fabric/pkg/peer/mocks"
	"github.com/trustbloc/sidetree-fabric/pkg/role"
)

//go:generate counterfeiter -o ../mocks/peerconfig.gen.go --fake-name PeerConfig . peerConfig

const (
	peer1 = "peer1.example.com"
	msp1  = "Org1MSP"
)

// Ensure that the roles are loaded
var _ = extroles.GetRoles()

func TestObserverController(t *testing.T) {
	peerCfg := &peermocks.PeerConfig{}
	peerCfg.PeerIDReturns(peer1)
	peerCfg.MSPIDReturns(msp1)

	observerCfg := config.Observer{Period: time.Second}
	providers := &observer.ClientProviders{}
	gossip := extmocks.NewMockGossipAdapter()
	gossip.Self(msp1, extmocks.NewMember(peer1, []byte("pkiid")))

	gossipProvider := &extmocks.GossipProvider{}
	gossipProvider.GetGossipServiceReturns(gossip)

	providers.Gossip = gossipProvider

	txnChan := make(chan gossipapi.TxMetadata)

	t.Run("Observer is started", func(t *testing.T) {
		rolesValue := make(map[extroles.Role]struct{})
		rolesValue[role.Observer] = struct{}{}
		extroles.SetRoles(rolesValue)
		defer func() {
			extroles.SetRoles(nil)
		}()

		m := newObserverController(channel1, peerCfg, observerCfg, providers, txnChan, coremocks.NewMockProtocolClientProvider())
		require.NotNil(t, m)

		require.NoError(t, m.Start())
		time.Sleep(100 * time.Millisecond)
		m.Stop()
	})

	t.Run("Observer is not started", func(t *testing.T) {
		rolesValue := make(map[extroles.Role]struct{})
		rolesValue[extroles.EndorserRole] = struct{}{}
		rolesValue[role.Resolver] = struct{}{}
		extroles.SetRoles(rolesValue)
		defer func() {
			extroles.SetRoles(nil)
		}()

		m := newObserverController(channel1, peerCfg, observerCfg, providers, txnChan, coremocks.NewMockProtocolClientProvider())
		require.NotNil(t, m)

		require.NoError(t, m.Start())
		time.Sleep(100 * time.Millisecond)
		m.Stop()
	})
}
