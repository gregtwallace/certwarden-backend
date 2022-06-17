package sqlite

import "database/sql"

// Funcs to transform payload types into null types (for db obj)

// boolToNullBool converts a *bool to a NullBool
func boolToNullBool(b *bool) sql.NullBool {
	var nullBool sql.NullBool

	if b == nil {
		nullBool.Valid = false
	} else {
		nullBool.Valid = true
		nullBool.Bool = *b
	}

	return nullBool
}

// intToNullInt32 converts an *int to a NullInt32
func intToNullInt32(i *int) sql.NullInt32 {
	var nullInt32 sql.NullInt32

	if i == nil {
		nullInt32.Valid = false
	} else {
		nullInt32.Valid = true
		nullInt32.Int32 = int32(*i)
	}

	return nullInt32
}

// stringToNullString converts a *string to a NullString
func stringToNullString(s *string) sql.NullString {
	var nullString sql.NullString

	if s == nil {
		nullString.Valid = false
	} else {
		nullString.Valid = true
		nullString.String = *s
	}

	return nullString
}
