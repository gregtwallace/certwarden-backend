package output

import (
	"encoding/json"
	"strings"
)

// RedactString removes the middle portion of a string and returns only the first and last
// characters separated by asterisks.
func RedactString(s string) string {
	// if s is less than or equal to 4 characters, do not redact; nothing sensitive
	// should ever be this small
	if len(s) <= 5 {
		return s
	}

	// return first 3 + asterisks + last 2
	return string(s[:2]) + "************" + string(s[len(s)-2:])
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

// TryUnredact attempts to 'fix' rs if the real value of rs contains redacted
// characteristics (e.g. contains a bunch of redact chars ***). It does this
// by comparing the redacted version of realVal against rs. If rs unredacted
// is equal to the redacted realVal, rs is set to the unredacted realVal. If
// not, rs is not modified.
func (rs *RedactedString) TryUnredact(realVal string) {
	realValAsRs := RedactedString(realVal)

	// check if rs is the same as redacted realVal
	if rs.Unredacted() == realValAsRs.Redacted() {
		*rs = realValAsRs
	}
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

// redactEnvironmentParam redacts a string but only the portion after the = sign
func redactEnvironmentParam(envVar string) string {
	// split on first instance of =
	vArr := strings.SplitN(envVar, "=", 2)

	// if no = was found, redact entire string
	if len(vArr) <= 1 {
		envVar = RedactString(envVar)
	} else {
		envVar = vArr[0] + "=" + RedactString(vArr[1])
	}

	return envVar
}

func (rep RedactedEnvironmentParams) Unredacted() []string {
	return []string(rep)
}

// Redacted returns a slice of redacted strings of RedactedEnvironmentParams
func (rep RedactedEnvironmentParams) Redacted() []string {
	if rep == nil {
		return nil
	}

	// make return slice
	ret := make([]string, len(rep))

	// populate redacted slice
	for i, v := range rep {
		ret[i] = redactEnvironmentParam(v)
	}

	return ret
}

// TryUnredact attempts to 'fix' rep.  It does this by checking each of the
// redacted verison of each possibleRealVals against each of the members of
// rep. If a match is found, the real value is substituted in for that element
// of rep. If no match is found, that element is not changed.
func (rep RedactedEnvironmentParams) TryUnredact(possibleRealVals []string) {
	for repIndex, envVar := range rep {
		for possRealValIndex, possRealVal := range possibleRealVals {
			// if envVar == redacted possible value, good to go and update rep element
			if envVar == redactEnvironmentParam(possRealVal) {
				// unredact the rep element
				rep[repIndex] = possRealVal

				// delete the matched real val from options to avoid duplicate in the
				// extremely unlikely event that it matches more than once
				possibleRealVals[possRealValIndex] = possibleRealVals[len(possibleRealVals)-1]
				possibleRealVals = possibleRealVals[:len(possibleRealVals)-1]

				// done with this envVar once matched
				break
			}
		}
	}
}

// MarshalJSON for RedactedEnvironmentParams first redacts the slice of string and
// then Marshals it
func (rep *RedactedEnvironmentParams) MarshalJSON() ([]byte, error) {
	return json.Marshal(rep.Redacted())
}
