package validation

import "time"

// time.go
var TimeNow = func() func() time.Time { return timeNow } // returns current timeNow function
var SetTimeNow = setTimeNow
