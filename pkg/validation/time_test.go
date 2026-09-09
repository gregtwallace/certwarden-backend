package validation_test

import (
	"certwarden-backend/pkg/validation"
	"testing"
	"time"
)

func TestSetTimeNow(t *testing.T) {
	setTiVal := time.Unix(888888, 0)

	// change time
	revert := validation.SetTimeNow(t, setTiVal)

	// check change
	ti := validation.TimeNow()()
	if !ti.Equal(setTiVal) {
		t.Errorf("time change didn't work, expected '%s' but got '%s'", setTiVal, ti)
	}

	// check revert to real func
	revert()
	ti = validation.TimeNow()()
	if ti.Equal(setTiVal) {
		t.Errorf("time change revert didn't work, expected current time but got '%s'", ti)
	}
}
