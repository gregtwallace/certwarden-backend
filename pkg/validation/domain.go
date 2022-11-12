package validation

import (
	"regexp"
)

// DomainValidRegex is the regex to confirm a domain is in the proper
// form.
const DomainValidRegex = `^(([A-Za-z0-9][A-Za-z0-9-]{0,61}\.)*([A-Za-z0-9][A-Za-z0-9-]{0,61}\.)[A-Za-z][A-Za-z0-9-]{0,61}[A-Za-z0-9])$`

// DomainValid returns true if the string is a validly formatted
// domain name
// https://tools.ietf.org/id/draft-liman-tld-names-00.html
// this is likely more inclusive than ACME server will permit
// TODO(?): restrict this further
func DomainValid(domain string) bool {
	return regexp.MustCompile(DomainValidRegex).MatchString(domain)
}
