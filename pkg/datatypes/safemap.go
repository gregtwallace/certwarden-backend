package datatypes

import (
	"errors"
	"strings"
	"sync"
)

// errors
var (
	errMapKeyDoesntExist       = errors.New("specified map key does not exist")
	errMapKeySuffixDoesntExist = errors.New("could not find map key satisfying suffix")
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
func (sm *SafeMap[V]) Read(key string) (V, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	// read data
	value, exists := sm.m[key]
	if !exists {
		return value, errMapKeyDoesntExist
	}

	return value, nil
}

// ReadSuffix searches the map for a key that is the Suffix of the specified
// string. If one is found, the key/value pair is returned. If no key satisfies
// this constraint, an error is returned.
func (sm *SafeMap[V]) ReadSuffix(s string) (key string, value V, err error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	for key = range sm.m {
		if strings.HasSuffix(s, key) {
			return key, sm.m[key], nil
		}
	}

	return "", value, errMapKeySuffixDoesntExist
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
		return errMapKeyDoesntExist
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

// Len returns the length of the map (i.e. the number of key/value pairs)
func (sm *SafeMap[V]) Len() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	return len(sm.m)
}

// CheckValuesForFunc runs the specified against each value in the map until
// one of the value returns true. If one returns true, true is returned. If
// none of the values return true, then false is returned.
func (sm *SafeMap[V]) CheckValuesForFunc(checkValueFunc func(value V) bool) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	// range through map and return true if any checkValueFunc returns true
	for _, v := range sm.m {
		if checkValueFunc(v) {
			return true
		}
	}

	return false
}
