package sqlite

import (
	"database/sql"
)

// Funcs to transform sql types into correspoinding pointer type

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
