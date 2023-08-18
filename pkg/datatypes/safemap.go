package datatypes

import (
	"errors"
	"sync"
)

// errors
var errMapElementDoesntExist = errors.New("specified map key does not exist")

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
func (sm *SafeMap[V]) Read(key string) (V, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	// read data
	value, exists := sm.m[key]
	if !exists {
		return value, errMapElementDoesntExist
	}

	return value, nil
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

// DeleteKey deletes the specified key from the map.
// If no such key exists, an error is returned.
func (sm *SafeMap[V]) DeleteKey(key string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// if key was not found, error
	_, exists := sm.m[key]
	if !exists {
		return errMapElementDoesntExist
	}

	// delete the key
	delete(sm.m, key)

	return nil
}

// DeleteFunc deletes any key/value pairs where the function passed in
// returns true
func (sm *SafeMap[V]) DeleteFunc(delFunc func(key string, value V) bool) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// range through map and delete if delFunc returns true
	for key, v := range sm.m {
		if delFunc(key, v) {
			delete(sm.m, key)
		}
	}
}
