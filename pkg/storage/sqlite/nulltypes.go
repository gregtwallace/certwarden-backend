package sqlite

import "database/sql"

// Funcs to transform payload types into null types (for db obj)
// as well as null type to regular type

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

// nullBoolToBool converts the nullbool to a bool pointer
func nullBoolToBool(nullBool sql.NullBool) *bool {
	if nullBool.Valid {
		b := new(bool)
		*b = nullBool.Bool

		return b
	}

	return nil
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

// NullInt32ToInt converts a NullInt32 into an int pointer
func nullInt32ToInt(nullInt sql.NullInt32) *int {
	if nullInt.Valid {
		i := new(int)
		*i = int(nullInt.Int32)

		return i
	}

	return nil
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

// nullStringToString converts the nullstring to a string pointer
func nullStringToString(nullString sql.NullString) *string {
	if nullString.Valid {
		s := new(string)
		*s = nullString.String

		return s
	}

	return nil
}
