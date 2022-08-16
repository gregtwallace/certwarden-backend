package datatypes

import (
	"errors"
	"sync"
)

// errors
var errMapElementDoesntExist = errors.New("specified map element does not exist")

// SafeMap is a map with a mutex
type SafeMap struct {
	Map map[string]interface{}
	sync.RWMutex
}

// newMap creates a new SafeMap
func NewSafeMap() *SafeMap {
	return &SafeMap{}
}

// Read returns the data from the specified elementName. If the element
// does not exist, an error is returned.
func (safeMap *SafeMap) Read(elementName string) (data interface{}, err error) {
	safeMap.RLock()
	defer safeMap.RUnlock()

	// read data
	data, exists := safeMap.Map[elementName]
	if !exists {
		return nil, errMapElementDoesntExist
	}

	return data, nil
}

// Add creates the named element and insert the specified data
// if the element already exists, true and the existing data are
// returned instead.
func (safeMap *SafeMap) Add(elementName string, data interface{}) (alreadyExists bool, existingData interface{}) {
	safeMap.Lock()
	defer safeMap.Unlock()

	// if element exists, return true and the existing data
	// check if already working, if so return the signal channel
	existingData, exists := safeMap.Map[elementName]
	if exists {
		return true, existingData
	}

	// if not, add the element and data
	safeMap.Map[elementName] = data

	return false, nil
}

// Delete deletes the specified elementName from the map.
// If no such element exists, an error is returned.
func (safeMap *SafeMap) Delete(elementName string) (err error) {
	safeMap.Lock()
	defer safeMap.Unlock()

	// if element was not found, error
	_, exists := safeMap.Map[elementName]
	if !exists {
		return errMapElementDoesntExist
	}

	// delete the element
	delete(safeMap.Map, elementName)

	return nil
}
