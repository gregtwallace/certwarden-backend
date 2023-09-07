package output

import (
	"encoding/json"
	"strings"
)

// RedactString removes the middle portion of a string and returns only the first and last
// characters separated by asterisks. if the key is less than or equal to 12 chars only
// asterisks are returned
func RedactString(s string) string {
	// if the identifier is less than 12 chars in length, return fully redacted
	// this should never happen but just in case to prevent credential logging
	if len(s) <= 12 {
		return "************"
	}

	// return first 2 + asterisks + last 1
	return string(s[:2]) + "************" + string(s[len(s)-1:])
}

// RedactedString is a string that will always be redacted when it is json
// marshalled but NOT when it is other types of marshalled
type RedactedString string

// Unredacted returns the complete string of RedactedString
func (rs *RedactedString) Unredacted() string {
	if rs == nil {
		return ""
	}

	return string(*rs)
}

// Redacted returns the redacted string of RedactedString
func (rs *RedactedString) Redacted() string {
	if rs == nil {
		return ""
	}

	return RedactString(string(*rs))
}

// MarshalJSON for redactedString first redacts the string and then Marshals it
func (rs *RedactedString) MarshalJSON() ([]byte, error) {
	return json.Marshal(rs.Redacted())
}

// RedactedEnvironmentParams is a slice of strings that will always be redacted when it is
// json marshalled but NOT when it is other types of marshalled. It only redacts the portion
// of each string following the first appearance of an = symbol. This is to cause
// redaction such that "SomeParam = 123456abcdefg" redacts to "SomeParam = 12**********g"
type RedactedEnvironmentParams []string

func (rep *RedactedEnvironmentParams) Unredacted() []string {
	return []string(*rep)
}

// Redacted returns a slice of redacted strings of RedactedEnvironmentParams
func (rep *RedactedEnvironmentParams) Redacted() []string {
	if rep == nil {
		return nil
	}

	// make return slice
	ret := make([]string, len(*rep))

	// populate redacted slice
	for i, v := range *rep {
		// split on first instance of =
		vArr := strings.SplitN(v, "=", 2)

		// if no = was found, redact entire string
		if len(vArr) <= 1 {
			ret[i] = RedactString(v)
		} else {
			ret[i] = vArr[0] + "=" + RedactString(vArr[1])
		}
	}

	return ret
}

// MarshalJSON for RedactedEnvironmentParams first redacts the slice of string and
// then Marshals it
func (rep *RedactedEnvironmentParams) MarshalJSON() ([]byte, error) {
	return json.Marshal(rep.Redacted())
}
