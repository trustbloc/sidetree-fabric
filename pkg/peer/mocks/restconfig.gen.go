// Code generated by counterfeiter. DO NOT EDIT.
package mocks

import (
	"sync"
)

type RestConfig struct {
	SidetreeListenURLStub        func() (string, error)
	sidetreeListenURLMutex       sync.RWMutex
	sidetreeListenURLArgsForCall []struct{}
	sidetreeListenURLReturns     struct {
		result1 string
		result2 error
	}
	sidetreeListenURLReturnsOnCall map[int]struct {
		result1 string
		result2 error
	}
	SidetreeListenPortStub        func() int
	sidetreeListenPortMutex       sync.RWMutex
	sidetreeListenPortArgsForCall []struct{}
	sidetreeListenPortReturns     struct {
		result1 int
	}
	sidetreeListenPortReturnsOnCall map[int]struct {
		result1 int
	}
	SidetreeTLSCertificateStub        func() string
	sidetreeTLSCertificateMutex       sync.RWMutex
	sidetreeTLSCertificateArgsForCall []struct{}
	sidetreeTLSCertificateReturns     struct {
		result1 string
	}
	sidetreeTLSCertificateReturnsOnCall map[int]struct {
		result1 string
	}
	SidetreeTLSKeyStub        func() string
	sidetreeTLSKeyMutex       sync.RWMutex
	sidetreeTLSKeyArgsForCall []struct{}
	sidetreeTLSKeyReturns     struct {
		result1 string
	}
	sidetreeTLSKeyReturnsOnCall map[int]struct {
		result1 string
	}
	SidetreeAPITokenStub        func(name string) string
	sidetreeAPITokenMutex       sync.RWMutex
	sidetreeAPITokenArgsForCall []struct {
		name string
	}
	sidetreeAPITokenReturns struct {
		result1 string
	}
	sidetreeAPITokenReturnsOnCall map[int]struct {
		result1 string
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *RestConfig) SidetreeListenURL() (string, error) {
	fake.sidetreeListenURLMutex.Lock()
	ret, specificReturn := fake.sidetreeListenURLReturnsOnCall[len(fake.sidetreeListenURLArgsForCall)]
	fake.sidetreeListenURLArgsForCall = append(fake.sidetreeListenURLArgsForCall, struct{}{})
	fake.recordInvocation("SidetreeListenURL", []interface{}{})
	fake.sidetreeListenURLMutex.Unlock()
	if fake.SidetreeListenURLStub != nil {
		return fake.SidetreeListenURLStub()
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fake.sidetreeListenURLReturns.result1, fake.sidetreeListenURLReturns.result2
}

func (fake *RestConfig) SidetreeListenURLCallCount() int {
	fake.sidetreeListenURLMutex.RLock()
	defer fake.sidetreeListenURLMutex.RUnlock()
	return len(fake.sidetreeListenURLArgsForCall)
}

func (fake *RestConfig) SidetreeListenURLReturns(result1 string, result2 error) {
	fake.SidetreeListenURLStub = nil
	fake.sidetreeListenURLReturns = struct {
		result1 string
		result2 error
	}{result1, result2}
}

func (fake *RestConfig) SidetreeListenURLReturnsOnCall(i int, result1 string, result2 error) {
	fake.SidetreeListenURLStub = nil
	if fake.sidetreeListenURLReturnsOnCall == nil {
		fake.sidetreeListenURLReturnsOnCall = make(map[int]struct {
			result1 string
			result2 error
		})
	}
	fake.sidetreeListenURLReturnsOnCall[i] = struct {
		result1 string
		result2 error
	}{result1, result2}
}

func (fake *RestConfig) SidetreeListenPort() int {
	fake.sidetreeListenPortMutex.Lock()
	ret, specificReturn := fake.sidetreeListenPortReturnsOnCall[len(fake.sidetreeListenPortArgsForCall)]
	fake.sidetreeListenPortArgsForCall = append(fake.sidetreeListenPortArgsForCall, struct{}{})
	fake.recordInvocation("SidetreeListenPort", []interface{}{})
	fake.sidetreeListenPortMutex.Unlock()
	if fake.SidetreeListenPortStub != nil {
		return fake.SidetreeListenPortStub()
	}
	if specificReturn {
		return ret.result1
	}
	return fake.sidetreeListenPortReturns.result1
}

func (fake *RestConfig) SidetreeListenPortCallCount() int {
	fake.sidetreeListenPortMutex.RLock()
	defer fake.sidetreeListenPortMutex.RUnlock()
	return len(fake.sidetreeListenPortArgsForCall)
}

func (fake *RestConfig) SidetreeListenPortReturns(result1 int) {
	fake.SidetreeListenPortStub = nil
	fake.sidetreeListenPortReturns = struct {
		result1 int
	}{result1}
}

func (fake *RestConfig) SidetreeListenPortReturnsOnCall(i int, result1 int) {
	fake.SidetreeListenPortStub = nil
	if fake.sidetreeListenPortReturnsOnCall == nil {
		fake.sidetreeListenPortReturnsOnCall = make(map[int]struct {
			result1 int
		})
	}
	fake.sidetreeListenPortReturnsOnCall[i] = struct {
		result1 int
	}{result1}
}

func (fake *RestConfig) SidetreeTLSCertificate() string {
	fake.sidetreeTLSCertificateMutex.Lock()
	ret, specificReturn := fake.sidetreeTLSCertificateReturnsOnCall[len(fake.sidetreeTLSCertificateArgsForCall)]
	fake.sidetreeTLSCertificateArgsForCall = append(fake.sidetreeTLSCertificateArgsForCall, struct{}{})
	fake.recordInvocation("SidetreeTLSCertificate", []interface{}{})
	fake.sidetreeTLSCertificateMutex.Unlock()
	if fake.SidetreeTLSCertificateStub != nil {
		return fake.SidetreeTLSCertificateStub()
	}
	if specificReturn {
		return ret.result1
	}
	return fake.sidetreeTLSCertificateReturns.result1
}

func (fake *RestConfig) SidetreeTLSCertificateCallCount() int {
	fake.sidetreeTLSCertificateMutex.RLock()
	defer fake.sidetreeTLSCertificateMutex.RUnlock()
	return len(fake.sidetreeTLSCertificateArgsForCall)
}

func (fake *RestConfig) SidetreeTLSCertificateReturns(result1 string) {
	fake.SidetreeTLSCertificateStub = nil
	fake.sidetreeTLSCertificateReturns = struct {
		result1 string
	}{result1}
}

func (fake *RestConfig) SidetreeTLSCertificateReturnsOnCall(i int, result1 string) {
	fake.SidetreeTLSCertificateStub = nil
	if fake.sidetreeTLSCertificateReturnsOnCall == nil {
		fake.sidetreeTLSCertificateReturnsOnCall = make(map[int]struct {
			result1 string
		})
	}
	fake.sidetreeTLSCertificateReturnsOnCall[i] = struct {
		result1 string
	}{result1}
}

func (fake *RestConfig) SidetreeTLSKey() string {
	fake.sidetreeTLSKeyMutex.Lock()
	ret, specificReturn := fake.sidetreeTLSKeyReturnsOnCall[len(fake.sidetreeTLSKeyArgsForCall)]
	fake.sidetreeTLSKeyArgsForCall = append(fake.sidetreeTLSKeyArgsForCall, struct{}{})
	fake.recordInvocation("SidetreeTLSKey", []interface{}{})
	fake.sidetreeTLSKeyMutex.Unlock()
	if fake.SidetreeTLSKeyStub != nil {
		return fake.SidetreeTLSKeyStub()
	}
	if specificReturn {
		return ret.result1
	}
	return fake.sidetreeTLSKeyReturns.result1
}

func (fake *RestConfig) SidetreeTLSKeyCallCount() int {
	fake.sidetreeTLSKeyMutex.RLock()
	defer fake.sidetreeTLSKeyMutex.RUnlock()
	return len(fake.sidetreeTLSKeyArgsForCall)
}

func (fake *RestConfig) SidetreeTLSKeyReturns(result1 string) {
	fake.SidetreeTLSKeyStub = nil
	fake.sidetreeTLSKeyReturns = struct {
		result1 string
	}{result1}
}

func (fake *RestConfig) SidetreeTLSKeyReturnsOnCall(i int, result1 string) {
	fake.SidetreeTLSKeyStub = nil
	if fake.sidetreeTLSKeyReturnsOnCall == nil {
		fake.sidetreeTLSKeyReturnsOnCall = make(map[int]struct {
			result1 string
		})
	}
	fake.sidetreeTLSKeyReturnsOnCall[i] = struct {
		result1 string
	}{result1}
}

func (fake *RestConfig) SidetreeAPIToken(name string) string {
	fake.sidetreeAPITokenMutex.Lock()
	ret, specificReturn := fake.sidetreeAPITokenReturnsOnCall[len(fake.sidetreeAPITokenArgsForCall)]
	fake.sidetreeAPITokenArgsForCall = append(fake.sidetreeAPITokenArgsForCall, struct {
		name string
	}{name})
	fake.recordInvocation("SidetreeAPIToken", []interface{}{name})
	fake.sidetreeAPITokenMutex.Unlock()
	if fake.SidetreeAPITokenStub != nil {
		return fake.SidetreeAPITokenStub(name)
	}
	if specificReturn {
		return ret.result1
	}
	return fake.sidetreeAPITokenReturns.result1
}

func (fake *RestConfig) SidetreeAPITokenCallCount() int {
	fake.sidetreeAPITokenMutex.RLock()
	defer fake.sidetreeAPITokenMutex.RUnlock()
	return len(fake.sidetreeAPITokenArgsForCall)
}

func (fake *RestConfig) SidetreeAPITokenArgsForCall(i int) string {
	fake.sidetreeAPITokenMutex.RLock()
	defer fake.sidetreeAPITokenMutex.RUnlock()
	return fake.sidetreeAPITokenArgsForCall[i].name
}

func (fake *RestConfig) SidetreeAPITokenReturns(result1 string) {
	fake.SidetreeAPITokenStub = nil
	fake.sidetreeAPITokenReturns = struct {
		result1 string
	}{result1}
}

func (fake *RestConfig) SidetreeAPITokenReturnsOnCall(i int, result1 string) {
	fake.SidetreeAPITokenStub = nil
	if fake.sidetreeAPITokenReturnsOnCall == nil {
		fake.sidetreeAPITokenReturnsOnCall = make(map[int]struct {
			result1 string
		})
	}
	fake.sidetreeAPITokenReturnsOnCall[i] = struct {
		result1 string
	}{result1}
}

func (fake *RestConfig) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.sidetreeListenURLMutex.RLock()
	defer fake.sidetreeListenURLMutex.RUnlock()
	fake.sidetreeListenPortMutex.RLock()
	defer fake.sidetreeListenPortMutex.RUnlock()
	fake.sidetreeTLSCertificateMutex.RLock()
	defer fake.sidetreeTLSCertificateMutex.RUnlock()
	fake.sidetreeTLSKeyMutex.RLock()
	defer fake.sidetreeTLSKeyMutex.RUnlock()
	fake.sidetreeAPITokenMutex.RLock()
	defer fake.sidetreeAPITokenMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *RestConfig) recordInvocation(key string, args []interface{}) {
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
