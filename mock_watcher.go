package confl

import (
	"errors"
	"reflect"
	"sync"
)

// mockWatcher watcher for test
type mockWatcher struct {
	c       interface{}
	oCopyed interface{}
	nCopyL  sync.RWMutex
	nCopyed interface{}
	hookL   sync.Mutex
	hooks   []Hook
	change  chan struct{}
}

// NewMockWatcher new watcher with pointer of config stucture
func NewMockWatcher(c interface{}) (*mockWatcher, error) {
	if reflect.ValueOf(c).IsNil() {
		return nil, errors.New("need none nil config structure")
	}
	return &mockWatcher{
		c:      c,
		hooks:  []Hook{},
		change: make(chan struct{}),
	}, nil
}

// Config Watcher interface method Config
func (m *mockWatcher) Config() interface{} {
	m.nCopyL.RLock()
	defer m.nCopyL.RUnlock()
	return m.nCopyed
}

// Close Watcher interface method Close
func (m *mockWatcher) Close() error {
	return nil
}

// AddHook Watcher interface method AddHook
func (m *mockWatcher) AddHook(hs ...Hook) {
	m.hookL.Lock()
	m.hooks = append(m.hooks, hs...)
	m.hookL.Unlock()
}

// OnError Watcher interface method OnError
func (m *mockWatcher) OnError(f func(error)) {
}

// Watch Watcher interface method Watch
func (m *mockWatcher) Watch() {
	for range m.change {
		for _, hook := range m.hooks {
			hook(m.oCopyed, m.Config())
		}
	}
}

// Trigger trigger the watch
func (m *mockWatcher) Trigger() {
	m.change <- struct{}{}
}
