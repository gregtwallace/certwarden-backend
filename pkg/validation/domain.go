package validation

import (
	"regexp"
	"strconv"
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

// DomainAndPortValid validates if the string is a valid fqdn followed by a colon followed
// by a valid port number
func DomainAndPortValid(domain string) bool {
	clientAddrSplit := strings.Split(domain, ":")

	// bad split
	if len(clientAddrSplit) != 1 && len(clientAddrSplit) != 2 {
		return false
	}

	// validate port (if exists)
	if len(clientAddrSplit) == 2 {
		portNumb, err := strconv.Atoi(clientAddrSplit[1])
		if err != nil || portNumb < 1 || portNumb > 65535 {
			return false
		}
	}

	// validate domain
	return regexp.MustCompile(DomainValidRegex).MatchString(clientAddrSplit[0])
}

// HttpsUrlValid returns true if the string contains only valid URL characters
func HttpsUrlValid(url string) bool {
	if !strings.HasPrefix(url, "https://") {
		return false
	}

	return regexp.MustCompile(URLValidRegex).MatchString(url)
}
