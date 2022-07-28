package sqlite

import (
	"database/sql"
	"encoding/json"
	"legocerthub-backend/pkg/acme"
	"strings"
)

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

// sliceToCommaNullString converts a slice of strings into a single string
// separated by commas
func sliceToCommaNullString(ss []string) sql.NullString {
	var nullString sql.NullString

	if ss == nil {
		nullString.Valid = false
	} else {
		nullString.Valid = true
		nullString.String = strings.Join(ss, ",")
	}

	return nullString
}

// commaNullStringToSlice converts a string that is a comma separated list
// of strings into a slice of strings
func commaNullStringToSlice(nullString sql.NullString) []string {
	if nullString.Valid {
		// if the string isn't empty, parse it
		if nullString.String != "" {
			s := nullStringToString(nullString)
			slice := strings.Split(*s, ",")
			return slice
		} else {
			// empty string = empty
			return nil
		}
	}

	return nil
}

// acmeErrorToNullString converts an acme.Error into a json string and then
// converts that into a NullString
func acmeErrorToNullString(acmeErr *acme.Error) (nullString sql.NullString) {
	if acmeErr == nil {
		nullString.Valid = false
		return nullString
	}

	acmeErrBytes, err := json.Marshal(acmeErr)
	if err != nil {
		nullString.Valid = false
		return nullString
	}

	nullString.Valid = true
	nullString.String = string(acmeErrBytes)
	return nullString
}

// nullStringToAcmeError converts a json NullString into an acme.Error
// object
func nullStringToAcmeError(nullString sql.NullString) (acmeErr *acme.Error) {
	if nullString.Valid {
		// if valid, unmarshal
		acmeErr := new(acme.Error)
		err := json.Unmarshal([]byte(nullString.String), acmeErr)
		if err != nil {
			// if unmarshal fails, not valid, return nil
			return nil
		}
		return acmeErr
	}

	return nil
}
