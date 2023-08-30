package output

// RedactString removes the middle portion of a string and returns only the first and last
// characters separated by asterisks. if the key is less than or equal to 12 chars only
// asterisks are returned
func RedactString(id string) string {
	// if the identifier is less than 12 chars in length, return fully redacted
	// this should never happen but just in case to prevent credential logging
	if len(id) <= 12 {
		return "************"
	}

	// return first 2 + asterisks + last 1
	return string(id[:2]) + "************" + string(id[len(id)-1:])
}
