package validation

import (
	"regexp"
)

// EmailValidRegex is the regex to confirm an email is in the proper
// form.
const EmailValidRegex = `^[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,4}$`

// EmailValid returns true if the string contains a
// validly formatted email address
func EmailValid(email string) bool {
	// valid email regex
	emailRegex := regexp.MustCompile(EmailValidRegex)

	return emailRegex.MatchString(email)
}

// EmailValidOrBlank returns true if the email is blank or
// contains a valid email format
func EmailValidOrBlank(email string) bool {
	// blank check
	if email == "" {
		return true
	}

	// regex check
	return EmailValid(email)
}
