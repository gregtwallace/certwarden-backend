package safemap

import (
	"sync"
)

// SafeMap is a map with a mutex
type SafeMap[V any] struct {
	m  map[string]V
	mu sync.RWMutex
}

// NewSafeMap creates a new SafeMap
func NewSafeMap[V any]() *SafeMap[V] {
	return &SafeMap[V]{
		m: make(map[string]V),
	}
}

// Read returns the value from the specified key. If the key
// does not exist, an error is returned.
func (sm *SafeMap[V]) Read(key string) (V, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	// read data
	value, exists := sm.m[key]

	return value, exists
}

// Add creates the named key and inserts the specified value.
// If the key already exists, true and the existing value are
// returned instead.
func (sm *SafeMap[V]) Add(key string, value V) (bool, V) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// if key exists, return true and the existing value
	exisingValue, exists := sm.m[key]
	if exists {
		return true, exisingValue
	}

	// if not, add the key and value
	sm.m[key] = value

	return false, value
}

// DeleteFunc deletes any key/value pairs where the function passed in
// returns true; it returns true if something was deleted, it returns
// false if nothing was deleted
func (sm *SafeMap[V]) DeleteFunc(delFunc func(key string, value V) bool) bool {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// range through map and delete if delFunc returns true
	didDelete := false
	for key, val := range sm.m {
		if delFunc(key, val) {
			didDelete = true
			delete(sm.m, key)
		}
	}

	return didDelete
}
