/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package sidetreesvc

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	extmocks "github.com/trustbloc/fabric-peer-ext/pkg/mocks"
	extroles "github.com/trustbloc/fabric-peer-ext/pkg/roles"
	"github.com/trustbloc/sidetree-fabric/pkg/observer"
	"github.com/trustbloc/sidetree-fabric/pkg/role"
)

func TestObserverController(t *testing.T) {
	bp := extmocks.NewBlockPublisherProvider()

	providers := &observer.Providers{
		BlockPublisher: bp,
	}

	t.Run("Observer role", func(t *testing.T) {
		rolesValue := make(map[extroles.Role]struct{})
		rolesValue[role.Observer] = struct{}{}
		extroles.SetRoles(rolesValue)
		defer func() {
			extroles.SetRoles(nil)
		}()

		o := newObserverController(channel1, providers)
		require.NotNil(t, o)

		require.NoError(t, o.Start())
		time.Sleep(100 * time.Millisecond)
		o.Stop()
	})

	t.Run("No observer role", func(t *testing.T) {
		rolesValue := make(map[extroles.Role]struct{})
		extroles.SetRoles(rolesValue)
		defer func() {
			extroles.SetRoles(nil)
		}()

		o := newObserverController(channel1, providers)
		require.NotNil(t, o)

		require.NoError(t, o.Start())
		o.Stop()
	})
}
