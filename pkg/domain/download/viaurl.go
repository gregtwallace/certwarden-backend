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

	// trim any prepended `/`
	return strings.TrimPrefix(apiKey, "/")
}
