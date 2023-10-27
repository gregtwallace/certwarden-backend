package download

import (
	"net/http"
	"strings"
)

// apiKeyHeaderNames are the potential header names that will be checked
// for the api key
var apiKeyHeaderNames = []string{
	"apiKey",
	"X-API-Key",
}

// getApiKeyFromHeader gets the api key from approved headers and also
// modifies ResponseWriter to include the Vary header re: api key
func getApiKeyFromHeader(w http.ResponseWriter, r *http.Request) (apiKeyHeaderValue string) {
	// all api key fields should be included in Vary
	varyCanonicalVals := []string{}
	for _, varyVal := range apiKeyHeaderNames {
		// use canonical names
		varyCanonicalVals = append(varyCanonicalVals, http.CanonicalHeaderKey(varyVal))
	}
	varyValsString := strings.Join(varyCanonicalVals, ", ")

	// add to Vary
	w.Header().Add("Vary", varyValsString)

	// find api key value
	for _, headerName := range apiKeyHeaderNames {
		// set value of header found
		if r.Header.Get(headerName) != "" {
			return r.Header.Get(headerName)
		}
	}

	return ""
}
