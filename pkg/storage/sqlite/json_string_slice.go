package sqlite

import (
	"encoding/json"
)

// jsonStringSlice is a string type in storage that is a json formatted
// array of strings
type jsonStringSlice string

// transform JSS into string slice
func (jss jsonStringSlice) toSlice() []string {
	if jss == "" {
		return []string{}
	}

	strSlice := []string{}
	err := json.Unmarshal([]byte(jss), &strSlice)
	if err != nil {
		return []string{}
	}

	return strSlice
}

// makeCommaJoinedString creates a JSS from a slice of strings
func makeJsonStringSlice(stringSlice []string) jsonStringSlice {
	if len(stringSlice) == 0 {
		return "[]"
	}

	jss, err := json.Marshal(stringSlice)
	if err != nil {
		return "[]"
	}

	return jsonStringSlice(jss)
}
