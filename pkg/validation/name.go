package validation

import (
	"regexp"
)

// NameValidRegex is the regex to confirm a name is in the proper
// form.
const NameValidRegex = `[^-_.~A-z0-9]|[\^]`

// NameValid true if the specified name is acceptable. To be valid
// the name must only contain symbols - _ . ~ letters and numbers,
// and name cannot be blank (len <= 0)
func NameValid(name string) bool {
	regex := regexp.MustCompile(NameValidRegex)

	invalid := regex.Match([]byte(name))
	if invalid || len(name) <= 0 {
		return false
	}
	return true
}
