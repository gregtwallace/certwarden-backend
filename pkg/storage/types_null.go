package storage

import (
	"database/sql"
	"time"
)

// Funcs to transform sql types into correspoinding pointer type

// nullInt64UnixToTime converts a NullInt64 into an a time.Time pointer
func nullInt64UnixToTime(nullInt sql.NullInt64) *time.Time {
	if nullInt.Valid {
		t := time.Unix(int64(nullInt.Int64), 0)

		return &t
	}

	return nil
}

// nullStringToString converts the nullstring to a string pointer
func nullStringToString(nullString sql.NullString) *string {
	if nullString.Valid {
		s := new(string)
		*s = nullString.String

		return s
	}

	return nil
}
