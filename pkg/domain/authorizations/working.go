package authorizations

import (
	"errors"
	"sync"
)

var ErrWorkingFailedToRemove = errors.New("auth working cannot remove non existent url")

// working is a struct to track which auths are being worked on. urls is a map of
// URLs being worked on and corresponding signal channels. The signal channel is used
// to block if there are subsequent calls to work an auth that is already queued in working and
// subsequently when the channel is closed this signals the auth is done being worked
// to the other waiting threads.
type working struct {
	urls map[string]chan struct{}
	mu   sync.Mutex
}

// newWorking creates the working struct for tracking orders being worked on
func newWorking() *working {
	u := make(map[string]chan struct{})

	return &working{
		urls: u,
	}
}

// add adds the specified auth url to the working slice (which tracks
// auths currently being worked).
// Returns true if the url already exists along with a signal channel.
// Returns false if does not already exists and creates a new signal channel, no channel
// is returned.
func (working *working) add(newAuthUrl string) (alreadyExists bool, signal chan struct{}) {
	working.mu.Lock()
	defer working.mu.Unlock()

	// check if already working, if so return the signal channel
	signal, exists := working.urls[newAuthUrl]
	if exists {
		return true, signal
	}

	// if not, add the url and make a signal channel for it
	working.urls[newAuthUrl] = make(chan struct{})

	return false, nil
}

// remove removes the specified order ID from the working slice.
// This is done after the fulfillment routine completes
func (working *working) remove(removeAuthUrl string) (err error) {
	working.mu.Lock()
	defer working.mu.Unlock()

	// if url was not found, error
	signal, exists := working.urls[removeAuthUrl]
	if !exists {
		return ErrWorkingFailedToRemove
	}

	// close the signal channel and remove the url
	delete(working.urls, removeAuthUrl)
	close(signal)

	return nil
}
