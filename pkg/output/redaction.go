package output

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
