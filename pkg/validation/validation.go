package validation

import (
	"errors"
	"regexp"
)

var (
	// domain
	ErrDomainBad     = errors.New("bad domain or subject name")
	ErrDomainMissing = errors.New("missing domain or subject")

	// order
	ErrOrderMismatch     = errors.New("order cert id does not match cert")
	ErrOrderValid        = errors.New("order is already valid")
	ErrOrderInvalid      = errors.New("order is invalid and cannot be retried")
	ErrOrderNotRevocable = errors.New("order (cert) cannot be revoked")
)

// DomainValid returns true if the string is a validly formatted
// domain name
// https://tools.ietf.org/id/draft-liman-tld-names-00.html
// this is likely more inclusive than ACME server will permit
// TODO(?): restrict this further
func DomainValid(domain string) bool {
	// valid regex
	emailRegex, err := regexp.Compile(`^(([A-Za-z0-9][A-Za-z0-9-]{0,61}\.)*([A-Za-z0-9][A-Za-z0-9-]{0,61}\.)[A-Za-z][A-Za-z0-9-]{0,61}[A-Za-z0-9])$`)
	if err != nil {
		// should never happen
		return false
	}

	return emailRegex.MatchString(domain)
}
