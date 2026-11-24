package storage

import (
	"database/sql"
	"time"
)

// Funcs to transform sql types into corresponding pointer type or vice versa

// nullInt64UnixToTime converts a NullInt64 into an a time.Time pointer
func nullInt64UnixToTime(nullInt sql.NullInt64) *time.Time {
	if nullInt.Valid {
		return new(time.Unix(int64(nullInt.Int64), 0))
	}

	return nil
}

// timePointerToNullInt64 converts a time.Time pointer to a NullInt64 with the
// int value set to the Unix time value of the time.Time pointer
func timePointerToNullInt64(tPtr *time.Time) sql.NullInt64 {
	// not valid (nil)
	if tPtr == nil {
		return sql.NullInt64{Valid: false}
	}

	// is valid
	return sql.NullInt64{
		Valid: true,
		Int64: tPtr.Unix(),
	}
}

// nullStringToString converts the nullstring to a string pointer
func nullStringToString(nullString sql.NullString) *string {
	if nullString.Valid {
		return new(nullString.String)
	}

	return nil
}
