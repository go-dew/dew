package dew

import (
	"reflect"
	"sync"
)

// syncMap is a structure that holds a map of handlers.
type syncMap struct {
	kv map[reflect.Type]any
	mu sync.RWMutex
}

// load returns the value stored in the map.
func (m *syncMap) load(key reflect.Type) (value any, ok bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	value, ok = m.kv[key]
	return
}

// store stores the value in the map.
func (m *syncMap) store(key reflect.Type, value any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.kv[key] = value
}
