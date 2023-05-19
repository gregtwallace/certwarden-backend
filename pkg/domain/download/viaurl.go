package download

import (
	"strings"

	"github.com/julienschmidt/httprouter"
)

// getApiKeyFromParams parses the apiKey wild card param to ensure only
// the apiKey is returned. The remainder of the param is discarded.
func getApiKeyFromParams(params httprouter.Params) (apiKey string) {
	// get the wildcard apikey param
	apiKey = params.ByName("apiKey")

	// split apiKey at slashes (/)
	pieces := strings.Split(apiKey, "/")

	// the param always starts with a slash, so second piece is the apiKey
	return pieces[1]
}
