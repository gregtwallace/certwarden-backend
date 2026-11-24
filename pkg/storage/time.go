package storage

import (
	"testing"
	"time"
)

var timeNow = time.Now

// Test related code below.

// NOTE: This is deliberately 'storage' and not 'storage_test'. This enables
// using this exported override function in other test packages, without
// exporting the package's global timeNow var.

func Now() time.Time { return timeNow() }

// SetTimeNow sets the storage pkg global timeNow variable to the specified
// time, and returns a function that reverts the timeNow global var to the
// time.Now function
// The function doesn't use *testing.T, but instead requires it to prevent
// accidental use in a non-test function.
func SetTimeNow(_ *testing.T, ti time.Time) (revertToDefaultTimeNow func()) {
	cancelFunc := func() {
		timeNow = time.Now
	}

	timeNow = func() time.Time {
		return ti
	}

	return cancelFunc
}
