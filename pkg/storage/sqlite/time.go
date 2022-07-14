package sqlite

import (
	"database/sql"
	"time"
)

// timeNow() returns unix time as a NullInt32
func timeNow() (unixTime sql.NullInt32) {
	timeNow := int(time.Now().Unix())

	return intToNullInt32(&timeNow)
}
