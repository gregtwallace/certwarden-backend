package storage

import "time"

// timeNow() returns unix time as a NullInt32
// TODO: Remove
func timeNow() (unixTime int) {
	return int(time.Now().Unix())
}
