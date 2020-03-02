/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package operationqueue

import (
	"sync"

	"github.com/hyperledger/fabric/common/flogging"
	"github.com/trustbloc/sidetree-core-go/pkg/batch/cutter"
)

var logger = flogging.MustGetLogger("sidetree_opqueue")

type key struct {
	channelID string
	namespace string
}

// Provider manages operation queues
type Provider struct {
	baseDir string
	queues  map[key]*LevelDBQueue
	mutex   sync.RWMutex
}

type peerConfig interface {
	LevelDBOpQueueBasePath() string
}

// NewProvider returns a new Operation LevelDBQueue provider
func NewProvider(cfg peerConfig) *Provider {
	logger.Infof("Creating Sidetree operation queue provider")

	return &Provider{
		baseDir: cfg.LevelDBOpQueueBasePath(),
		queues:  make(map[key]*LevelDBQueue),
	}
}

// Create returns the operation queue for the given channel
func (p *Provider) Create(channelID, namespace string) (cutter.OperationQueue, error) {
	k := key{
		channelID: channelID,
		namespace: namespace,
	}

	p.mutex.RLock()
	q, ok := p.queues[k]
	p.mutex.RUnlock()

	if !ok {
		p.mutex.Lock()
		defer p.mutex.Unlock()

		var err error
		q, err = newLevelDBQueue(channelID, namespace, p.baseDir)
		if err != nil {
			return nil, err
		}

		p.queues[k] = q
	}

	return q, nil
}

// Close closes all databases
func (p *Provider) Close() {
	logger.Info("Closing operation queues...")

	p.mutex.RLock()
	defer p.mutex.RUnlock()

	for _, q := range p.queues {
		q.Close()
	}
}
