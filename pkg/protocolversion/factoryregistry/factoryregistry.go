/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package factoryregistry

import (
	"fmt"
	"sync"

	"github.com/hyperledger/fabric/common/flogging"
	"github.com/trustbloc/sidetree-core-go/pkg/api/cas"
	"github.com/trustbloc/sidetree-core-go/pkg/api/protocol"
	"github.com/trustbloc/sidetree-fabric/pkg/rest/sidetreehandler"

	ctxcommon "github.com/trustbloc/sidetree-fabric/pkg/context/common"
	"github.com/trustbloc/sidetree-fabric/pkg/protocolversion/common"
)

var logger = flogging.MustGetLogger("sidetree_peer")

type factory interface {
	Create(p protocol.Protocol, casClient cas.Client, opStore ctxcommon.OperationStore, docType sidetreehandler.DocumentType) (protocol.Version, error)
}

var mutex sync.RWMutex
var factories = make(map[string]factory)

// Registry implements a protocol version factory registry
type Registry struct {
}

// New returns a new protocol version factory Registry
func New() *Registry {
	logger.Info("Creating protocol version factory Registry")

	return &Registry{}
}

// CreateProtocolVersion creates a new protocol version using the given version, protocol and providers
func (m *Registry) CreateProtocolVersion(version string, p protocol.Protocol, casClient cas.Client, opStore ctxcommon.OperationStore, docType sidetreehandler.DocumentType) (protocol.Version, error) {
	v, err := m.resolveFactory(version)
	if err != nil {
		return nil, err
	}

	logger.Infof("Creating protocol version [%s]", version)

	return v.Create(p, casClient, opStore, docType)
}

// Register registers a protocol factory for a given version
func Register(version string, factory factory) {
	mutex.Lock()
	defer mutex.Unlock()

	if _, ok := factories[version]; ok {
		panic(fmt.Errorf("protocol version factory [%s] already registered", version))
	}

	logger.Infof("Registering protocol version factory [%s]", version)

	factories[version] = factory
}

func (m *Registry) resolveFactory(version string) (factory, error) {
	mutex.RLock()
	defer mutex.RUnlock()

	for v, f := range factories {
		if common.Version(v).Matches(version) {
			return f, nil
		}
	}

	return nil, fmt.Errorf("protocol version factory for version [%s] not found", version)
}
