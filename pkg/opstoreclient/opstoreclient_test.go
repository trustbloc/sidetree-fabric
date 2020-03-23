/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package opstoreclient

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/trustbloc/sidetree-fabric/pkg/mocks"
)

const (
	channel1 = "channel1"
	ns1      = "did:sidetree"
)

func TestProvider(t *testing.T) {
	dcasProvider := &mocks.DCASClientProvider{}

	p := NewProvider(dcasProvider)
	require.NotNil(t, p)

	s := p.Get(channel1, ns1)
	require.NotNil(t, s)
}
