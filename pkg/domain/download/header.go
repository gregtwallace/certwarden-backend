package download

import (
	"net/http"
	"strings"
)

// apiKeyHeaderNames are the potential header names that will be checked
// for the api key
var apiKeyHeaderNames = []string{
	"X-API-Key",
	"apiKey",
}

// getApiKeyFromHeader gets the api key from approved headers and also
// modifies ResponseWriter to include the Vary header re: api key
func getApiKeyFromHeader(w http.ResponseWriter, r *http.Request) (apiKeyHeaderValue string) {
	// all api key fields should be included in Vary
	varyVals := strings.Join(apiKeyHeaderNames, ", ")

	// if Vary is already set, include prior value in new value
	existingVary := w.Header().Get("Vary")
	if existingVary != "" {
		varyVals += ", " + existingVary
	}

	// set Vary
	w.Header().Set("Vary", varyVals)

	// find api key value
	for _, headerName := range apiKeyHeaderNames {
		// set value of header found
		if r.Header.Get(headerName) != "" {
			return r.Header.Get(headerName)
		}
	}

	return ""
}
