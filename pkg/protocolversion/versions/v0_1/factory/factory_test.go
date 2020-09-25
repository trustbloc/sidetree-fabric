/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package factory

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/trustbloc/sidetree-core-go/pkg/api/protocol"

	"github.com/trustbloc/sidetree-fabric/pkg/common"
	"github.com/trustbloc/sidetree-fabric/pkg/mocks"
)

func TestFactory_Create(t *testing.T) {
	f := New()
	require.NotNil(t, f)

	p := protocol.Protocol{}

	casClient := &mocks.CasClient{}
	opStore := &mocks.OperationStore{}

	t.Run("DID doc type", func(t *testing.T) {
		pv, err := f.Create("", p, casClient, opStore, common.DIDDocType)
		require.NoError(t, err)
		require.NotNil(t, pv)
	})

	t.Run("FileIndex doc type", func(t *testing.T) {
		pv, err := f.Create("", p, casClient, opStore, common.FileIndexType)
		require.NoError(t, err)
		require.NotNil(t, pv)
	})

	t.Run("Invalid doc type", func(t *testing.T) {
		pv, err := f.Create("", p, casClient, opStore, "invalid")
		require.EqualError(t, err, "unsupported document type: [invalid]")
		require.Nil(t, pv)
	})
}
