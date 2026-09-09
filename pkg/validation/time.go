package validation

import (
	"testing"
	"time"
)

var timeNow = time.Now

// setTimeNow sets the pkg global timeNow variable to the specified time, and returns
// a function that reverts the timeNow global var to the time.Now function
func setTimeNow(t *testing.T, ti time.Time) (revertToDefaultTimeNow func()) {
	t.Helper()

	cancelFunc := func() {
		timeNow = time.Now
	}

	timeNow = func() time.Time {
		return ti
	}

	return cancelFunc
}
