/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package sidetreesvc

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	extroles "github.com/trustbloc/fabric-peer-ext/pkg/roles"

	"github.com/trustbloc/sidetree-fabric/pkg/peer/config"
	"github.com/trustbloc/sidetree-fabric/pkg/peer/mocks"
	"github.com/trustbloc/sidetree-fabric/pkg/role"
)

//go:generate counterfeiter -o ../mocks/batchcontext.gen.go --fake-name BatchContext github.com/trustbloc/sidetree-core-go/pkg/batch.Context
//go:generate counterfeiter -o ../mocks/sidetreeconfigservice.gen.go --fake-name SidetreeConfigService ../config SidetreeService

const (
	namespace = "did:sidetree"
)

// Ensure that the roles are loaded
var _ = extroles.GetRoles()

func TestBatchWriter(t *testing.T) {
	rolesValue := make(map[extroles.Role]struct{})
	rolesValue[role.BatchWriter] = struct{}{}
	extroles.SetRoles(rolesValue)
	defer func() {
		extroles.SetRoles(nil)
	}()

	ctx := &mocks.BatchContext{}

	t.Run("Success", func(t *testing.T) {
		cfgService := &mocks.SidetreeConfigService{}
		cfgService.LoadSidetreeReturns(config.Sidetree{BatchWriterTimeout: time.Second}, nil)

		bw, err := newBatchWriter(channel1, namespace, ctx, cfgService)
		require.NoError(t, err)
		require.NotNil(t, bw)

		require.NoError(t, bw.Start())
		time.Sleep(100 * time.Millisecond)
		bw.Stop()
	})

	t.Run("sidetreeService error", func(t *testing.T) {
		errExpected := errors.New("injected sidetreeCfgService service error")
		cfgService := &mocks.SidetreeConfigService{}
		cfgService.LoadSidetreeReturns(config.Sidetree{}, errExpected)

		bw, err := newBatchWriter(channel1, namespace, ctx, cfgService)
		require.Error(t, err)
		require.Contains(t, err.Error(), errExpected.Error())
		require.Nil(t, bw)
	})
}
