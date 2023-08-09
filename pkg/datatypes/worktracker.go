package datatypes

import (
	"errors"
	"sync"
)

var errWTFailedToRemove = errors.New("work tracker cannot remove non existent item")

// WorkTracker is a struct to store names of items being worked on and corresponding signal
// channels that will be closed when a specific item is removed from the work tracker.
type WorkTracker struct {
	items map[string]chan struct{}
	sync.Mutex
}

// NewWorkTracker creates a WorkTracker
func NewWorkTracker() *WorkTracker {
	return &WorkTracker{
		items: make(map[string]chan struct{}),
	}
}

// Add adds the specified item to the WorkTracker (if it doesn't already exist) and creates
// a signal channel for it.
// If the item already exists, Add instead returns alreadyExists true and the existing
// signal channel for that item.
func (wt *WorkTracker) Add(item string) (alreadyExists bool, signal chan struct{}) {
	// lock for reading and possible modification
	wt.Lock()
	defer wt.Unlock()

	// check if item exists and if so return true and the signal channel for the item
	signal, exists := wt.items[item]
	if exists {
		return true, signal
	}

	// if item does not exist, add it, make a signal channel, and return false with the new
	// channel
	wt.items[item] = make(chan struct{})

	return false, wt.items[item]
}

// Remove closes the signal channel for the specified item and deletes it from the WorkTracker.
// If the item does not exist, in the WorkTracker, an error is returned.
func (wt *WorkTracker) Remove(item string) (err error) {
	// lock for reading and possible modification
	wt.Lock()
	defer wt.Unlock()

	// get item's channel
	signal, exists := wt.items[item]
	// it item was not found, error
	if !exists {
		return errWTFailedToRemove
	}

	// it item was found, close the signal channel and remove the item
	close(signal)
	delete(wt.items, item)

	return nil
}
