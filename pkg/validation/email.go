package validation

import (
	"regexp"
	"strings"
)

// EmailUsernameRegex is the regex to confirm an email username is in the proper
// form.
const emailUsernameRegex = `^[A-Za-z0-9][A-Za-z0-9-_.+]{0,62}[A-Za-z0-9]$`

// invalidConsecutiveSpecialRegex matches if the username contains two special
// chars in a row, which means it is not valid
const invalidConsecutiveSpecialRegex = `[-_.+]{2,}`

// EmailValid returns true if the string contains a validly formatted email address
func EmailValid(email string) bool {
	// split on @ and validate username and domain
	// also confirms exactly 1 @ symbol
	emailPieces := strings.Split(email, "@")
	if len(emailPieces) != 2 {
		return false
	}
	username := emailPieces[0]
	domain := emailPieces[1]

	// validate username
	if !regexp.MustCompile(emailUsernameRegex).MatchString(username) ||
		regexp.MustCompile(invalidConsecutiveSpecialRegex).MatchString(username) {
		return false
	}

	// validate domain
	if !DomainValid(domain, false) {
		return false
	}

	return true
}

// EmailValidOrBlank returns true if the email is blank or
// contains a valid email format
func EmailValidOrBlank(email string) bool {
	return email == "" || EmailValid(email)
}
