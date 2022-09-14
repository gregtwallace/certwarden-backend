package download

import "errors"

var (
	errBlankApiKey = errors.New("no apikey found")
	errWrongApiKey = errors.New("apikey is incorrect")

	errApiKeyFromUrlDisallowed = errors.New("apikey found in url but not allowed")

	errNoPem = errors.New("pem is blank")
)
