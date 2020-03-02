/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package operationqueue

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/trustbloc/sidetree-core-go/pkg/batch/cutter"
	"github.com/trustbloc/sidetree-fabric/pkg/context/operationqueue/mocks"
)

//go:generate counterfeiter -o ./mocks/peerconfig.gen.go --fake-name PeerConfig . peerConfig

const (
	channel_x = "channel_x"
	channel_y = "channel_y"
)

var (
	levelDBBasePath = filepath.Join(os.TempDir(), "sidetree_op_queue")
)

func TestProvider(t *testing.T) {
	defer func() {
		if err := os.RemoveAll(levelDBBasePath); err != nil {
			t.Errorf("Error removing temp dir [%s]: %s", levelDBBasePath, err)
		}
	}()

	peerConfig := &mocks.PeerConfig{}
	peerConfig.LevelDBOpQueueBasePathReturns(levelDBBasePath)

	p := NewProvider(peerConfig)
	require.NotNil(t, p)

	q1, err := p.Create(channel_x, namespace1)
	require.NoError(t, err)
	require.NotNil(t, q1)
	defer cleanup(q1)

	q2, err := p.Create(channel_y, namespace1)
	require.NoError(t, err)
	require.NotNil(t, q2)
	defer cleanup(q2)

	require.False(t, q1 == q2)

	p.Close()
}

func cleanup(q cutter.OperationQueue) {
	q.(*LevelDBQueue).Close()
	if err := q.(*LevelDBQueue).Drop(); err != nil {
		logger.Infof("Error dropping queue: %s", err)
	}
}
