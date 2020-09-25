/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package factoryregistry

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/trustbloc/sidetree-core-go/pkg/api/protocol"
	coremocks "github.com/trustbloc/sidetree-core-go/pkg/mocks"

	"github.com/trustbloc/sidetree-fabric/pkg/common"
	"github.com/trustbloc/sidetree-fabric/pkg/mocks"
	frmocks "github.com/trustbloc/sidetree-fabric/pkg/protocolversion/factoryregistry/mocks"
)

//go:generate counterfeiter -o ./mocks/protocolfactory.gen.go --fake-name ProtocolFactory . factory

func TestRegistry(t *testing.T) {
	const version = "0.1"

	p := protocol.Protocol{}

	f := &frmocks.ProtocolFactory{}
	f.CreateReturns(&coremocks.ProtocolVersion{}, nil)

	require.NotPanics(t, func() { Register(version, f) })
	require.PanicsWithError(t, "protocol version factory [0.1] already registered", func() { Register(version, f) })

	r := New()

	casClient := &mocks.CasClient{}
	opStore := &mocks.OperationStore{}
	docType := common.DIDDocType

	pv, err := r.CreateProtocolVersion(version, p, casClient, opStore, docType)
	require.NoError(t, err)
	require.NotNil(t, pv)

	pv, err = r.CreateProtocolVersion("99", p, casClient, opStore, docType)
	require.EqualError(t, err, "protocol version factory for version [99] not found")
	require.Nil(t, pv)
}
