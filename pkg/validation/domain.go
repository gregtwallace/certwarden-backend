package validation

import (
	"regexp"
	"strings"
)

const DomainValidRegex = `^(([A-Za-z0-9][A-Za-z0-9-]{0,61}\.)*([A-Za-z0-9][A-Za-z0-9-]{0,61}\.)[A-Za-z][A-Za-z0-9-]{0,61}[A-Za-z0-9])$`
const URLValidRegex = `^[A-Za-z0-9-_.~!#$&'()*+,/:;=?@%[\]]*$`

// DomainValid returns true if the string is a validly formatted
// domain name
// https://tools.ietf.org/id/draft-liman-tld-names-00.html
// this is likely more inclusive than ACME server will permit
// TODO(?): restrict this further
func DomainValid(domain string, wildOk bool) bool {
	// if wildcard is allowed (for certs it is allowed per RFC 8555 7.1.3)
	if wildOk {
		// if string prefix is wildcard ("*."), remove it and then validate the remainder
		// if the prefix is not *. this call is a no-op
		domain = strings.TrimPrefix(domain, "*.")
	}

	return regexp.MustCompile(DomainValidRegex).MatchString(domain)
}

// HttpsUrlValid returns true if the string contains only valid URL characters
func HttpsUrlValid(url string) bool {
	if !strings.HasPrefix(url, "https://") {
		return false
	}

	return regexp.MustCompile(URLValidRegex).MatchString(url)
}
