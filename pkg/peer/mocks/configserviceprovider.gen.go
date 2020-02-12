// Code generated by counterfeiter. DO NOT EDIT.
package mocks

import (
	"sync"

	ledgerconfig "github.com/trustbloc/fabric-peer-ext/pkg/config/ledgerconfig/config"
)

type ConfigServiceProvider struct {
	ForChannelStub        func(channelID string) ledgerconfig.Service
	forChannelMutex       sync.RWMutex
	forChannelArgsForCall []struct {
		channelID string
	}
	forChannelReturns struct {
		result1 ledgerconfig.Service
	}
	forChannelReturnsOnCall map[int]struct {
		result1 ledgerconfig.Service
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *ConfigServiceProvider) ForChannel(channelID string) ledgerconfig.Service {
	fake.forChannelMutex.Lock()
	ret, specificReturn := fake.forChannelReturnsOnCall[len(fake.forChannelArgsForCall)]
	fake.forChannelArgsForCall = append(fake.forChannelArgsForCall, struct {
		channelID string
	}{channelID})
	fake.recordInvocation("ForChannel", []interface{}{channelID})
	fake.forChannelMutex.Unlock()
	if fake.ForChannelStub != nil {
		return fake.ForChannelStub(channelID)
	}
	if specificReturn {
		return ret.result1
	}
	return fake.forChannelReturns.result1
}

func (fake *ConfigServiceProvider) ForChannelCallCount() int {
	fake.forChannelMutex.RLock()
	defer fake.forChannelMutex.RUnlock()
	return len(fake.forChannelArgsForCall)
}

func (fake *ConfigServiceProvider) ForChannelArgsForCall(i int) string {
	fake.forChannelMutex.RLock()
	defer fake.forChannelMutex.RUnlock()
	return fake.forChannelArgsForCall[i].channelID
}

func (fake *ConfigServiceProvider) ForChannelReturns(result1 ledgerconfig.Service) {
	fake.ForChannelStub = nil
	fake.forChannelReturns = struct {
		result1 ledgerconfig.Service
	}{result1}
}

func (fake *ConfigServiceProvider) ForChannelReturnsOnCall(i int, result1 ledgerconfig.Service) {
	fake.ForChannelStub = nil
	if fake.forChannelReturnsOnCall == nil {
		fake.forChannelReturnsOnCall = make(map[int]struct {
			result1 ledgerconfig.Service
		})
	}
	fake.forChannelReturnsOnCall[i] = struct {
		result1 ledgerconfig.Service
	}{result1}
}

func (fake *ConfigServiceProvider) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.forChannelMutex.RLock()
	defer fake.forChannelMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *ConfigServiceProvider) recordInvocation(key string, args []interface{}) {
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