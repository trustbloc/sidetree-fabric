/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package factory

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/trustbloc/sidetree-core-go/pkg/api/protocol"

	"github.com/trustbloc/sidetree-fabric/pkg/mocks"
	"github.com/trustbloc/sidetree-fabric/pkg/rest/sidetreehandler"
)

func TestFactory_Create(t *testing.T) {
	f := New()
	require.NotNil(t, f)

	p := protocol.Protocol{}

	casClient := &mocks.CasClient{}
	opStore := &mocks.OperationStore{}

	t.Run("DID doc type", func(t *testing.T) {
		pv, err := f.Create(p, casClient, opStore, sidetreehandler.DIDDocType)
		require.NoError(t, err)
		require.NotNil(t, pv)
	})

	t.Run("FileIndex doc type", func(t *testing.T) {
		pv, err := f.Create(p, casClient, opStore, sidetreehandler.FileIndexType)
		require.NoError(t, err)
		require.NotNil(t, pv)
	})

	t.Run("Invalid doc type", func(t *testing.T) {
		pv, err := f.Create(p, casClient, opStore, "invalid")
		require.EqualError(t, err, "unsupported document type: [invalid]")
		require.Nil(t, pv)
	})
}
