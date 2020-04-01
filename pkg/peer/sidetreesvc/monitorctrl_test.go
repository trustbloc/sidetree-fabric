/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package sidetreesvc

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	extroles "github.com/trustbloc/fabric-peer-ext/pkg/roles"

	"github.com/trustbloc/sidetree-fabric/pkg/mocks"
	"github.com/trustbloc/sidetree-fabric/pkg/observer/monitor"
	"github.com/trustbloc/sidetree-fabric/pkg/peer/config"
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

func TestMonitorController(t *testing.T) {
	peerCfg := &peermocks.PeerConfig{}
	peerCfg.PeerIDReturns(peer1)
	peerCfg.MSPIDReturns(msp1)

	monitorCfg := config.Monitor{Period: time.Second}
	providers := &monitor.ClientProviders{}

	t.Run("Monitor is started", func(t *testing.T) {
		rolesValue := make(map[extroles.Role]struct{})
		rolesValue[extroles.CommitterRole] = struct{}{}
		rolesValue[role.Resolver] = struct{}{}
		extroles.SetRoles(rolesValue)
		defer func() {
			extroles.SetRoles(nil)
		}()

		m := newMonitorController(channel1, peerCfg, monitorCfg, providers, &mocks.OperationStoreProvider{})
		require.NotNil(t, m)

		require.NoError(t, m.Start())
		time.Sleep(100 * time.Millisecond)
		m.Stop()
	})

	t.Run("Monitor is not started", func(t *testing.T) {
		rolesValue := make(map[extroles.Role]struct{})
		rolesValue[extroles.EndorserRole] = struct{}{}
		rolesValue[role.Resolver] = struct{}{}
		extroles.SetRoles(rolesValue)
		defer func() {
			extroles.SetRoles(nil)
		}()

		m := newMonitorController(channel1, peerCfg, monitorCfg, providers, &mocks.OperationStoreProvider{})
		require.NotNil(t, m)

		require.NoError(t, m.Start())
		time.Sleep(100 * time.Millisecond)
		m.Stop()
	})
}
