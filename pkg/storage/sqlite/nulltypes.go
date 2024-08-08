package sqlite

import (
	"database/sql"
	"time"
)

// Funcs to transform sql types into correspoinding pointer type

// nullInt32UnixToTime converts a NullInt32 into an a time.Time pointer
func nullInt32UnixToTime(nullInt sql.NullInt32) *time.Time {
	if nullInt.Valid {
		t := time.Unix(int64(nullInt.Int32), 0)

		return &t
	}

	return nil
}

// NullInt32ToInt converts a NullInt32 into an int pointer
func nullInt32ToInt(nullInt sql.NullInt32) *int {
	if nullInt.Valid {
		i := new(int)
		*i = int(nullInt.Int32)

		return i
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
