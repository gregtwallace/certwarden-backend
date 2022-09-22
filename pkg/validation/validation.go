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

	// order
	ErrOrderMismatch     = errors.New("order cert id does not match cert")
	ErrOrderValid        = errors.New("order is already valid")
	ErrOrderInvalid      = errors.New("order is invalid and cannot be retried")
	ErrOrderNotRevocable = errors.New("order (cert) cannot be revoked")
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

// NameValid true if the specified name is acceptable. To be valid
// the name must only contain symbols - _ . ~ letters and numbers,
// and name cannot be blank (len <= 0)
func NameValid(name string) bool {
	regex, err := regexp.Compile(`[^-_.~A-z0-9]|[\^]`)
	if err != nil {
		// should never happen
		return false
	}

	invalid := regex.Match([]byte(name))
	if invalid || len(name) <= 0 {
		return false
	}
	return true
}

// EmailValid returns true if the string contains a
// validly formatted email address
func EmailValid(email string) bool {
	// valid email regex
	emailRegex := regexp.MustCompile(`^[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,4}$`)

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

	// valid regex
	emailRegex := regexp.MustCompile(`^(([A-Za-z0-9][A-Za-z0-9-]{0,61}\.)*([A-Za-z0-9][A-Za-z0-9-]{0,61}\.)[A-Za-z][A-Za-z0-9-]{0,61}[A-Za-z0-9])$`)
	isGood := emailRegex.MatchString(*domain)
	if isGood {
		return nil
	}

	return ErrDomainBad
}
