package sqlite

import (
	"time"
)

// timeNow() returns unix time as a NullInt32
func timeNow() (unixTime int) {
	return int(time.Now().Unix())
}
