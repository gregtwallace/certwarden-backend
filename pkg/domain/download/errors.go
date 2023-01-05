package download

import "errors"

var (
	errBlankApiKey = errors.New("no apikey found")
	errWrongApiKey = errors.New("apikey is incorrect")

	errApiKeyFromUrlDisallowed = errors.New("apikey found in url but not allowed")

	errApiDisabled = errors.New("download via api is disabled")

	errNoPem = errors.New("pem is blank")
)
