package validation

import "net/url"

// IsValidHTTPSURL returns true if the specified string parses as a URL and uses the scheme
// https and has a non-empty host
func IsValidHTTPSURL(rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}

	return u.Scheme == "https" && u.Host != ""
}
