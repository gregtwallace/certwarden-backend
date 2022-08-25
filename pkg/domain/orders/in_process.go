package orders

import (
	"errors"
	"sync"
)

var (
	ErrOrderAlreadyProcessing  = errors.New("order is already being processed")
	ErrInProcessFailedToRemove = errors.New("inprocess cannot remove non existent id")
)

// inProcess holds a list of orders being worked
type inProcess struct {
	ids []int
	mu  sync.Mutex
}

// newInProcess creates the inProcess struct for tracking orders being worked on
func newInProcess() *inProcess {
	return &inProcess{
		ids: []int{},
	}
}

// add adds the specified order ID to the inProcess slice (which tracks
// orders currently being worked)
func (ip *inProcess) add(orderId int) (err error) {
	ip.mu.Lock()
	defer ip.mu.Unlock()

	// check if already inProcess
	for i := range ip.ids {
		if ip.ids[i] == orderId {
			return ErrOrderAlreadyProcessing
		}
	}

	// if not, add
	ip.ids = append(ip.ids, orderId)
	return nil
}

// remove removes the specified order ID from the inProcess slice.
// This is done after the fulfillment routine completes
func (ip *inProcess) remove(orderId int) (err error) {
	ip.mu.Lock()
	defer ip.mu.Unlock()

	// find index
	i := -2
	for i = range ip.ids {
		if ip.ids[i] == orderId {
			break
		}
	}

	// if id was not found, error
	if i == -2 {
		return ErrInProcessFailedToRemove
	}

	// move last element to location of element being removed
	ip.ids[i] = ip.ids[len(ip.ids)-1]
	ip.ids = ip.ids[:len(ip.ids)-1]

	return nil
}
