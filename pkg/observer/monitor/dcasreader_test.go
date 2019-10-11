/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package monitor

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/trustbloc/sidetree-fabric/pkg/observer/common"
	obmocks "github.com/trustbloc/sidetree-fabric/pkg/observer/mocks"
	"github.com/trustbloc/sidetree-fabric/pkg/observer/monitor/mocks"
)

func TestSidetreeDCASReader_Read(t *testing.T) {
	dcasProvider := &mocks.DCASClientProvider{}
	dcasClient := obmocks.NewMockDCASClient()
	dcasProvider.ForChannelReturns(dcasClient)

	r := NewSidetreeDCASReader(channel1, dcasProvider)
	require.NotNil(t, r)

	expectedValue := []byte("some value")
	key, err := dcasClient.Put(common.SidetreeNs, common.SidetreeColl, expectedValue)
	require.NoError(t, err)

	value, err := r.Read(key)
	require.NoError(t, err)
	require.Equal(t, expectedValue, value)
}
