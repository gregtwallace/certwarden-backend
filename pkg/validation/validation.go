package validation

import (
	"errors"
	"regexp"
)

const newId = -1

var (
	// id
	ErrIdBad      = errors.New("bad id")
	ErrIdMissing  = errors.New("missing id")
	ErrIdMismatch = errors.New("id param and payload mismatch")

	// name
	ErrNameBad     = errors.New("bad name")
	ErrNameMissing = errors.New("missing name")
	ErrNameInUse   = errors.New("name already in use")

	// email
	ErrEmailBad     = errors.New("bad email")
	ErrEmailMissing = errors.New("missing email")

	// key
	ErrKeyBadOption = errors.New("invalid key option")
	ErrKeyBad       = errors.New("bad private key")

	// domain
	ErrDomainBad     = errors.New("bad domain or subject name")
	ErrDomainMissing = errors.New("missing domain or subject")
)

// IsIdNew returns an error if the id isn't the specified new
// id value, or if the id value isn't nil (unspecified)
func IsIdNew(id *int) error {
	if id != nil && *id != newId {
		return ErrIdBad
	}

	return nil
}

// IsIdExisting verifies id is specified, is not the new id and is greater
// than or equal to 0.  This is the first building block for id validation.
// Others include matching param to payload, and doing storage queries.
func IsIdExisting(idPayload *int) error {
	// verify payload id isn't nil
	if idPayload == nil {
		return ErrIdMissing
	}

	// check that the id is not the new id
	if *idPayload == newId {
		return ErrIdBad
	}

	// id must be >= 0
	if *idPayload >= 0 {
		return nil
	}

	return ErrIdBad
}

// IsIdExistingMatch implements IsIdExisting (not nil, non-new, >= 0) but
// also includes param and payload match.
func IsIdExistingMatch(idParam int, idPayload *int) error {
	// verify the payload is valid
	err := IsIdExisting(idPayload)
	if err != nil {
		return err
	}

	// check the payload and the URI match
	if *idPayload != idParam {
		return ErrIdMismatch
	}

	return nil
}

// IsNameValid returns an error if not valid, nil if valid
// to be valid: must only contain symbols - _ . ~ letters and numbers
// name is also not allowed to be blank (len <= 0)
func IsNameValid(namePayload *string) error {
	// error if not specified
	if namePayload == nil {
		return ErrNameMissing
	}

	regex, err := regexp.Compile("[^-_.~A-z0-9]|[\\^]")
	if err != nil {
		// should never happen
		return err
	}

	invalid := regex.Match([]byte(*namePayload))
	if invalid || len(*namePayload) <= 0 {
		return ErrNameBad
	}
	return nil
}

// IsEmailValid returns an error if not valid, nil if valid
// blank is not permissible
func IsEmailValid(emailPayload *string) error {
	// nil check, error
	if emailPayload == nil {
		return ErrEmailMissing
	}

	// valid email regex
	emailRegex := regexp.MustCompile(`^[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,4}$`)
	isGood := emailRegex.MatchString(*emailPayload)
	if isGood {
		return nil
	}

	return ErrEmailBad
}

// IsEmailValidOrBlank returns an error if not valid, nil if valid
// to be valid: must be either blank or an email address format
func IsEmailValidOrBlank(emailPayload *string) (err error) {
	// Check if email is valid (regex check)
	err = IsEmailValid(emailPayload)
	if err != nil {
		// blank is permissible
		if *emailPayload == "" {
			return nil
		}

		return err
	}

	return ErrEmailBad
}

// IsDomainValid returns an error if not valid, nil if valid
// to be valid, regex is based off of:
// https://tools.ietf.org/id/draft-liman-tld-names-00.html
// this is likely more inclusive than ACME server will permit
// TODO(?): restrict this further
func IsDomainValid(domain *string) (err error) {
	// nil check, error
	if domain == nil {
		return ErrDomainMissing
	}

	// valid email regex
	emailRegex := regexp.MustCompile(`^([A-Za-z][A-Za-z0-9-]{0,61}[A-Za-z0-9]\.)*[A-Za-z][A-Za-z0-9-]{0,61}[A-Za-z0-9]$`)
	isGood := emailRegex.MatchString(*domain)
	if isGood {
		return nil
	}

	return ErrDomainBad
}
