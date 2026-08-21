package storage

import (
	"testing"
	"time"
)

// NOTE: This is deliberately 'storage' and not 'storage_test'. This enables
// using this exported override function in other test packages, without
// exporting the package's global timeNow var.

// SetTimeNow sets the storage pkg global timeNow variable to the specified
// time, and returns a function that reverts the timeNow global var to the
// time.Now function
func SetTimeNow(t time.Time) (revertToDefaultTimeNow func()) {
	cancelFunc := func() {
		timeNow = time.Now
	}

	timeNow = func() time.Time {
		return t
	}

	return cancelFunc
}

func TestSetTimeNow(t *testing.T) {
	setTiVal := time.Unix(888888, 0)

	// change time
	revert := SetTimeNow(setTiVal)

	// check change
	ti := timeNow()
	if !ti.Equal(setTiVal) {
		t.Errorf("time change didn't work, expected '%s' but got '%s'", setTiVal, ti)
	}

	// check revert to real func
	revert()
	ti = timeNow()
	if ti.Equal(setTiVal) {
		t.Errorf("time change revert didn't work, expected current time but got '%s'", ti)
	}
}
