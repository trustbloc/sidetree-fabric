// Code generated by counterfeiter. DO NOT EDIT.
package mocks

import (
	"sync"

	"github.com/trustbloc/sidetree-core-go/pkg/api/protocol"
)

type ProtocolProvider struct {
	ProtocolStub        func() protocol.Client
	protocolMutex       sync.RWMutex
	protocolArgsForCall []struct{}
	protocolReturns     struct {
		result1 protocol.Client
	}
	protocolReturnsOnCall map[int]struct {
		result1 protocol.Client
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *ProtocolProvider) Protocol() protocol.Client {
	fake.protocolMutex.Lock()
	ret, specificReturn := fake.protocolReturnsOnCall[len(fake.protocolArgsForCall)]
	fake.protocolArgsForCall = append(fake.protocolArgsForCall, struct{}{})
	fake.recordInvocation("Protocol", []interface{}{})
	fake.protocolMutex.Unlock()
	if fake.ProtocolStub != nil {
		return fake.ProtocolStub()
	}
	if specificReturn {
		return ret.result1
	}
	return fake.protocolReturns.result1
}

func (fake *ProtocolProvider) ProtocolCallCount() int {
	fake.protocolMutex.RLock()
	defer fake.protocolMutex.RUnlock()
	return len(fake.protocolArgsForCall)
}

func (fake *ProtocolProvider) ProtocolReturns(result1 protocol.Client) {
	fake.ProtocolStub = nil
	fake.protocolReturns = struct {
		result1 protocol.Client
	}{result1}
}

func (fake *ProtocolProvider) ProtocolReturnsOnCall(i int, result1 protocol.Client) {
	fake.ProtocolStub = nil
	if fake.protocolReturnsOnCall == nil {
		fake.protocolReturnsOnCall = make(map[int]struct {
			result1 protocol.Client
		})
	}
	fake.protocolReturnsOnCall[i] = struct {
		result1 protocol.Client
	}{result1}
}

func (fake *ProtocolProvider) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.protocolMutex.RLock()
	defer fake.protocolMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *ProtocolProvider) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}