package orders

import (
	"errors"
	"sync"
)

var (
	ErrOrderAlreadyWorking   = errors.New("order is already being processed")
	ErrWorkingFailedToRemove = errors.New("order working cannot remove non existent id")
)

// working holds a list of orders being worked
type working struct {
	ids []int
	mu  sync.Mutex
}

// newWorking creates the working struct for tracking orders being worked on
func newWorking() *working {
	return &working{
		ids: []int{},
	}
}

// add adds the specified order ID to the working slice (which tracks
// orders currently being worked)
func (working *working) add(orderId int) (err error) {
	working.mu.Lock()
	defer working.mu.Unlock()

	// check if already working
	for i := range working.ids {
		if working.ids[i] == orderId {
			return ErrOrderAlreadyWorking
		}
	}

	// if not, add
	working.ids = append(working.ids, orderId)
	return nil
}

// remove removes the specified order ID from the working slice.
// This is done after the fulfillment routine completes
func (working *working) remove(orderId int) (err error) {
	working.mu.Lock()
	defer working.mu.Unlock()

	// find index
	i := -2
	for i = range working.ids {
		if working.ids[i] == orderId {
			break
		}
	}

	// if id was not found, error
	if i == -2 {
		return ErrWorkingFailedToRemove
	}

	// move last element to location of element being removed
	working.ids[i] = working.ids[len(working.ids)-1]
	working.ids = working.ids[:len(working.ids)-1]

	return nil
}
