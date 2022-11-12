package validation

import (
	"regexp"
)

// NameValidRegex is the regex to confirm a name is in the proper
// form. (Note: if match is found, name is INVALID)
const NameValidRegex = `[^-_.~A-z0-9]|[\^]`

// NameValid true if the specified name is acceptable. To be valid
// the name must only contain symbols - _ . ~ letters and numbers,
// and name cannot be blank (len <= 0)
func NameValid(name string) bool {
	// length
	if len(name) <= 0 {
		return false
	}

	// validate (if this matches, it is INVALID)
	return !(regexp.MustCompile(NameValidRegex).Match([]byte(name)))
}
