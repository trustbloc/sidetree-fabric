/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package role

import (
	"testing"

	"github.com/stretchr/testify/require"
	extroles "github.com/trustbloc/fabric-peer-ext/pkg/roles"
)

func TestIsResolver(t *testing.T) {
	require.False(t, IsResolver())

	restore := setRoles(Resolver)
	defer restore()

	require.True(t, IsResolver())
}

func TestIsBatchWriter(t *testing.T) {
	require.False(t, IsBatchWriter())

	restore := setRoles(BatchWriter)
	defer restore()

	require.True(t, IsBatchWriter())
}

func TestIsObserver(t *testing.T) {
	require.False(t, IsObserver())

	t.Run("Active Observer", func(t *testing.T) {
		restore := setRoles(Observer)
		defer restore()

		require.True(t, IsObserver())
	})

	t.Run("Standby Observer", func(t *testing.T) {
		restore := setRoles(ObserverStandby)
		defer restore()

		require.True(t, IsObserver())
	})
}

func setRoles(roles ...extroles.Role) (restore func()) {
	rolesValue := make(map[extroles.Role]struct{})
	for _, r := range roles {
		rolesValue[r] = struct{}{}
	}

	extroles.SetRoles(rolesValue)

	restore = func() {
		extroles.SetRoles(nil)
	}

	return
}
