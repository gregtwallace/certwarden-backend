package utils

import (
	"errors"
	"regexp"
)

const newId = -1

// IsIdNew returns an error if the id isn't the specified new
// id value, or if the id value isn't nil (unspecified)
func IsIdNew(id *int) error {
	if id != nil && *id != newId {
		return errors.New("invalid new id")
	}

	return nil
}

// IsIdExisting verifies id is specified, is not the new id and is greater
// than or equal to 0.  This is the first building block for id validation.
// Others include matching param to payload, and doing storage queries.
func IsIdExisting(idPayload *int) error {
	// verify payload id isn't nil
	if idPayload == nil {
		return errors.New("must specify id in payload")
	}

	// check that the id is not the new id
	if *idPayload == newId {
		return errors.New("existing id must not be new id")
	}

	// id must be >= 0
	if *idPayload >= 0 {
		return nil
	}

	return errors.New("invalid id")
}

// IsIdExistingMatch implements IsIdExisting but also includes param and
// payload match.
func IsIdExistingMatch(idParam int, idPayload *int) error {
	// verify the payload is valid
	err := IsIdExisting(idPayload)
	if err != nil {
		return err
	}

	// check the payload and the URI match
	if *idPayload != idParam {
		return errors.New("id param/payload mismatch")
	}

	return nil
}

// IsNameValid returns an error if not valid, nil if valid
// to be valid: must only contain symbols - _ . ~ letters and numbers
// name is also not allowed to be blank (len <= 0)
func IsNameValid(name string) error {
	regex, err := regexp.Compile("[^-_.~A-z0-9]|[\\^]")
	if err != nil {
		return err
	}

	invalid := regex.Match([]byte(name))
	if invalid || len(name) <= 0 {
		return errors.New("invalid name")
	}
	return nil
}

// IsEmailValidOrBlank returns an error if not valid, nil if valid
// to be valid: must be either blank or an email address format
func IsEmailValidOrBlank(email string) error {
	// blank is permissible
	if email == "" {
		return nil
	}

	// valid email regex
	emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	isGood := emailRegex.MatchString(email)
	if isGood {
		return nil
	}
	return errors.New("bad email address")
}
