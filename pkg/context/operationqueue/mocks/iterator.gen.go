// Code generated by counterfeiter. DO NOT EDIT.
package mocks

import (
	"sync"

	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/util"
)

type Iterator struct {
	FirstStub        func() bool
	firstMutex       sync.RWMutex
	firstArgsForCall []struct{}
	firstReturns     struct {
		result1 bool
	}
	firstReturnsOnCall map[int]struct {
		result1 bool
	}
	LastStub        func() bool
	lastMutex       sync.RWMutex
	lastArgsForCall []struct{}
	lastReturns     struct {
		result1 bool
	}
	lastReturnsOnCall map[int]struct {
		result1 bool
	}
	SeekStub        func(key []byte) bool
	seekMutex       sync.RWMutex
	seekArgsForCall []struct {
		key []byte
	}
	seekReturns struct {
		result1 bool
	}
	seekReturnsOnCall map[int]struct {
		result1 bool
	}
	NextStub        func() bool
	nextMutex       sync.RWMutex
	nextArgsForCall []struct{}
	nextReturns     struct {
		result1 bool
	}
	nextReturnsOnCall map[int]struct {
		result1 bool
	}
	PrevStub        func() bool
	prevMutex       sync.RWMutex
	prevArgsForCall []struct{}
	prevReturns     struct {
		result1 bool
	}
	prevReturnsOnCall map[int]struct {
		result1 bool
	}
	ReleaseStub            func()
	releaseMutex           sync.RWMutex
	releaseArgsForCall     []struct{}
	SetReleaserStub        func(releaser util.Releaser)
	setReleaserMutex       sync.RWMutex
	setReleaserArgsForCall []struct {
		releaser util.Releaser
	}
	ValidStub        func() bool
	validMutex       sync.RWMutex
	validArgsForCall []struct{}
	validReturns     struct {
		result1 bool
	}
	validReturnsOnCall map[int]struct {
		result1 bool
	}
	ErrorStub        func() error
	errorMutex       sync.RWMutex
	errorArgsForCall []struct{}
	errorReturns     struct {
		result1 error
	}
	errorReturnsOnCall map[int]struct {
		result1 error
	}
	KeyStub        func() []byte
	keyMutex       sync.RWMutex
	keyArgsForCall []struct{}
	keyReturns     struct {
		result1 []byte
	}
	keyReturnsOnCall map[int]struct {
		result1 []byte
	}
	ValueStub        func() []byte
	valueMutex       sync.RWMutex
	valueArgsForCall []struct{}
	valueReturns     struct {
		result1 []byte
	}
	valueReturnsOnCall map[int]struct {
		result1 []byte
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *Iterator) First() bool {
	fake.firstMutex.Lock()
	ret, specificReturn := fake.firstReturnsOnCall[len(fake.firstArgsForCall)]
	fake.firstArgsForCall = append(fake.firstArgsForCall, struct{}{})
	fake.recordInvocation("First", []interface{}{})
	fake.firstMutex.Unlock()
	if fake.FirstStub != nil {
		return fake.FirstStub()
	}
	if specificReturn {
		return ret.result1
	}
	return fake.firstReturns.result1
}

func (fake *Iterator) FirstCallCount() int {
	fake.firstMutex.RLock()
	defer fake.firstMutex.RUnlock()
	return len(fake.firstArgsForCall)
}

func (fake *Iterator) FirstReturns(result1 bool) {
	fake.FirstStub = nil
	fake.firstReturns = struct {
		result1 bool
	}{result1}
}

func (fake *Iterator) FirstReturnsOnCall(i int, result1 bool) {
	fake.FirstStub = nil
	if fake.firstReturnsOnCall == nil {
		fake.firstReturnsOnCall = make(map[int]struct {
			result1 bool
		})
	}
	fake.firstReturnsOnCall[i] = struct {
		result1 bool
	}{result1}
}

func (fake *Iterator) Last() bool {
	fake.lastMutex.Lock()
	ret, specificReturn := fake.lastReturnsOnCall[len(fake.lastArgsForCall)]
	fake.lastArgsForCall = append(fake.lastArgsForCall, struct{}{})
	fake.recordInvocation("Last", []interface{}{})
	fake.lastMutex.Unlock()
	if fake.LastStub != nil {
		return fake.LastStub()
	}
	if specificReturn {
		return ret.result1
	}
	return fake.lastReturns.result1
}

func (fake *Iterator) LastCallCount() int {
	fake.lastMutex.RLock()
	defer fake.lastMutex.RUnlock()
	return len(fake.lastArgsForCall)
}

func (fake *Iterator) LastReturns(result1 bool) {
	fake.LastStub = nil
	fake.lastReturns = struct {
		result1 bool
	}{result1}
}

func (fake *Iterator) LastReturnsOnCall(i int, result1 bool) {
	fake.LastStub = nil
	if fake.lastReturnsOnCall == nil {
		fake.lastReturnsOnCall = make(map[int]struct {
			result1 bool
		})
	}
	fake.lastReturnsOnCall[i] = struct {
		result1 bool
	}{result1}
}

func (fake *Iterator) Seek(key []byte) bool {
	var keyCopy []byte
	if key != nil {
		keyCopy = make([]byte, len(key))
		copy(keyCopy, key)
	}
	fake.seekMutex.Lock()
	ret, specificReturn := fake.seekReturnsOnCall[len(fake.seekArgsForCall)]
	fake.seekArgsForCall = append(fake.seekArgsForCall, struct {
		key []byte
	}{keyCopy})
	fake.recordInvocation("Seek", []interface{}{keyCopy})
	fake.seekMutex.Unlock()
	if fake.SeekStub != nil {
		return fake.SeekStub(key)
	}
	if specificReturn {
		return ret.result1
	}
	return fake.seekReturns.result1
}

func (fake *Iterator) SeekCallCount() int {
	fake.seekMutex.RLock()
	defer fake.seekMutex.RUnlock()
	return len(fake.seekArgsForCall)
}

func (fake *Iterator) SeekArgsForCall(i int) []byte {
	fake.seekMutex.RLock()
	defer fake.seekMutex.RUnlock()
	return fake.seekArgsForCall[i].key
}

func (fake *Iterator) SeekReturns(result1 bool) {
	fake.SeekStub = nil
	fake.seekReturns = struct {
		result1 bool
	}{result1}
}

func (fake *Iterator) SeekReturnsOnCall(i int, result1 bool) {
	fake.SeekStub = nil
	if fake.seekReturnsOnCall == nil {
		fake.seekReturnsOnCall = make(map[int]struct {
			result1 bool
		})
	}
	fake.seekReturnsOnCall[i] = struct {
		result1 bool
	}{result1}
}

func (fake *Iterator) Next() bool {
	fake.nextMutex.Lock()
	ret, specificReturn := fake.nextReturnsOnCall[len(fake.nextArgsForCall)]
	fake.nextArgsForCall = append(fake.nextArgsForCall, struct{}{})
	fake.recordInvocation("Next", []interface{}{})
	fake.nextMutex.Unlock()
	if fake.NextStub != nil {
		return fake.NextStub()
	}
	if specificReturn {
		return ret.result1
	}
	return fake.nextReturns.result1
}

func (fake *Iterator) NextCallCount() int {
	fake.nextMutex.RLock()
	defer fake.nextMutex.RUnlock()
	return len(fake.nextArgsForCall)
}

func (fake *Iterator) NextReturns(result1 bool) {
	fake.NextStub = nil
	fake.nextReturns = struct {
		result1 bool
	}{result1}
}

func (fake *Iterator) NextReturnsOnCall(i int, result1 bool) {
	fake.NextStub = nil
	if fake.nextReturnsOnCall == nil {
		fake.nextReturnsOnCall = make(map[int]struct {
			result1 bool
		})
	}
	fake.nextReturnsOnCall[i] = struct {
		result1 bool
	}{result1}
}

func (fake *Iterator) Prev() bool {
	fake.prevMutex.Lock()
	ret, specificReturn := fake.prevReturnsOnCall[len(fake.prevArgsForCall)]
	fake.prevArgsForCall = append(fake.prevArgsForCall, struct{}{})
	fake.recordInvocation("Prev", []interface{}{})
	fake.prevMutex.Unlock()
	if fake.PrevStub != nil {
		return fake.PrevStub()
	}
	if specificReturn {
		return ret.result1
	}
	return fake.prevReturns.result1
}

func (fake *Iterator) PrevCallCount() int {
	fake.prevMutex.RLock()
	defer fake.prevMutex.RUnlock()
	return len(fake.prevArgsForCall)
}

func (fake *Iterator) PrevReturns(result1 bool) {
	fake.PrevStub = nil
	fake.prevReturns = struct {
		result1 bool
	}{result1}
}

func (fake *Iterator) PrevReturnsOnCall(i int, result1 bool) {
	fake.PrevStub = nil
	if fake.prevReturnsOnCall == nil {
		fake.prevReturnsOnCall = make(map[int]struct {
			result1 bool
		})
	}
	fake.prevReturnsOnCall[i] = struct {
		result1 bool
	}{result1}
}

func (fake *Iterator) Release() {
	fake.releaseMutex.Lock()
	fake.releaseArgsForCall = append(fake.releaseArgsForCall, struct{}{})
	fake.recordInvocation("Release", []interface{}{})
	fake.releaseMutex.Unlock()
	if fake.ReleaseStub != nil {
		fake.ReleaseStub()
	}
}

func (fake *Iterator) ReleaseCallCount() int {
	fake.releaseMutex.RLock()
	defer fake.releaseMutex.RUnlock()
	return len(fake.releaseArgsForCall)
}

func (fake *Iterator) SetReleaser(releaser util.Releaser) {
	fake.setReleaserMutex.Lock()
	fake.setReleaserArgsForCall = append(fake.setReleaserArgsForCall, struct {
		releaser util.Releaser
	}{releaser})
	fake.recordInvocation("SetReleaser", []interface{}{releaser})
	fake.setReleaserMutex.Unlock()
	if fake.SetReleaserStub != nil {
		fake.SetReleaserStub(releaser)
	}
}

func (fake *Iterator) SetReleaserCallCount() int {
	fake.setReleaserMutex.RLock()
	defer fake.setReleaserMutex.RUnlock()
	return len(fake.setReleaserArgsForCall)
}

func (fake *Iterator) SetReleaserArgsForCall(i int) util.Releaser {
	fake.setReleaserMutex.RLock()
	defer fake.setReleaserMutex.RUnlock()
	return fake.setReleaserArgsForCall[i].releaser
}

func (fake *Iterator) Valid() bool {
	fake.validMutex.Lock()
	ret, specificReturn := fake.validReturnsOnCall[len(fake.validArgsForCall)]
	fake.validArgsForCall = append(fake.validArgsForCall, struct{}{})
	fake.recordInvocation("Valid", []interface{}{})
	fake.validMutex.Unlock()
	if fake.ValidStub != nil {
		return fake.ValidStub()
	}
	if specificReturn {
		return ret.result1
	}
	return fake.validReturns.result1
}

func (fake *Iterator) ValidCallCount() int {
	fake.validMutex.RLock()
	defer fake.validMutex.RUnlock()
	return len(fake.validArgsForCall)
}

func (fake *Iterator) ValidReturns(result1 bool) {
	fake.ValidStub = nil
	fake.validReturns = struct {
		result1 bool
	}{result1}
}

func (fake *Iterator) ValidReturnsOnCall(i int, result1 bool) {
	fake.ValidStub = nil
	if fake.validReturnsOnCall == nil {
		fake.validReturnsOnCall = make(map[int]struct {
			result1 bool
		})
	}
	fake.validReturnsOnCall[i] = struct {
		result1 bool
	}{result1}
}

func (fake *Iterator) Error() error {
	fake.errorMutex.Lock()
	ret, specificReturn := fake.errorReturnsOnCall[len(fake.errorArgsForCall)]
	fake.errorArgsForCall = append(fake.errorArgsForCall, struct{}{})
	fake.recordInvocation("Error", []interface{}{})
	fake.errorMutex.Unlock()
	if fake.ErrorStub != nil {
		return fake.ErrorStub()
	}
	if specificReturn {
		return ret.result1
	}
	return fake.errorReturns.result1
}

func (fake *Iterator) ErrorCallCount() int {
	fake.errorMutex.RLock()
	defer fake.errorMutex.RUnlock()
	return len(fake.errorArgsForCall)
}

func (fake *Iterator) ErrorReturns(result1 error) {
	fake.ErrorStub = nil
	fake.errorReturns = struct {
		result1 error
	}{result1}
}

func (fake *Iterator) ErrorReturnsOnCall(i int, result1 error) {
	fake.ErrorStub = nil
	if fake.errorReturnsOnCall == nil {
		fake.errorReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.errorReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *Iterator) Key() []byte {
	fake.keyMutex.Lock()
	ret, specificReturn := fake.keyReturnsOnCall[len(fake.keyArgsForCall)]
	fake.keyArgsForCall = append(fake.keyArgsForCall, struct{}{})
	fake.recordInvocation("Key", []interface{}{})
	fake.keyMutex.Unlock()
	if fake.KeyStub != nil {
		return fake.KeyStub()
	}
	if specificReturn {
		return ret.result1
	}
	return fake.keyReturns.result1
}

func (fake *Iterator) KeyCallCount() int {
	fake.keyMutex.RLock()
	defer fake.keyMutex.RUnlock()
	return len(fake.keyArgsForCall)
}

func (fake *Iterator) KeyReturns(result1 []byte) {
	fake.KeyStub = nil
	fake.keyReturns = struct {
		result1 []byte
	}{result1}
}

func (fake *Iterator) KeyReturnsOnCall(i int, result1 []byte) {
	fake.KeyStub = nil
	if fake.keyReturnsOnCall == nil {
		fake.keyReturnsOnCall = make(map[int]struct {
			result1 []byte
		})
	}
	fake.keyReturnsOnCall[i] = struct {
		result1 []byte
	}{result1}
}

func (fake *Iterator) Value() []byte {
	fake.valueMutex.Lock()
	ret, specificReturn := fake.valueReturnsOnCall[len(fake.valueArgsForCall)]
	fake.valueArgsForCall = append(fake.valueArgsForCall, struct{}{})
	fake.recordInvocation("Value", []interface{}{})
	fake.valueMutex.Unlock()
	if fake.ValueStub != nil {
		return fake.ValueStub()
	}
	if specificReturn {
		return ret.result1
	}
	return fake.valueReturns.result1
}

func (fake *Iterator) ValueCallCount() int {
	fake.valueMutex.RLock()
	defer fake.valueMutex.RUnlock()
	return len(fake.valueArgsForCall)
}

func (fake *Iterator) ValueReturns(result1 []byte) {
	fake.ValueStub = nil
	fake.valueReturns = struct {
		result1 []byte
	}{result1}
}

func (fake *Iterator) ValueReturnsOnCall(i int, result1 []byte) {
	fake.ValueStub = nil
	if fake.valueReturnsOnCall == nil {
		fake.valueReturnsOnCall = make(map[int]struct {
			result1 []byte
		})
	}
	fake.valueReturnsOnCall[i] = struct {
		result1 []byte
	}{result1}
}

func (fake *Iterator) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.firstMutex.RLock()
	defer fake.firstMutex.RUnlock()
	fake.lastMutex.RLock()
	defer fake.lastMutex.RUnlock()
	fake.seekMutex.RLock()
	defer fake.seekMutex.RUnlock()
	fake.nextMutex.RLock()
	defer fake.nextMutex.RUnlock()
	fake.prevMutex.RLock()
	defer fake.prevMutex.RUnlock()
	fake.releaseMutex.RLock()
	defer fake.releaseMutex.RUnlock()
	fake.setReleaserMutex.RLock()
	defer fake.setReleaserMutex.RUnlock()
	fake.validMutex.RLock()
	defer fake.validMutex.RUnlock()
	fake.errorMutex.RLock()
	defer fake.errorMutex.RUnlock()
	fake.keyMutex.RLock()
	defer fake.keyMutex.RUnlock()
	fake.valueMutex.RLock()
	defer fake.valueMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *Iterator) recordInvocation(key string, args []interface{}) {
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

var _ iterator.Iterator = new(Iterator)
