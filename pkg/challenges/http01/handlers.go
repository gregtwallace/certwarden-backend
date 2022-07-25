package http01

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// challengeHandler responds to the ACME http-01 challenge path. If the requested
// token exists in this service's tokens, the expected token content is sent back to
// the client. If the token is not in the service's tokens, a 404 reply is sent.
func (service *Service) challengeHandler(w http.ResponseWriter, r *http.Request) {
	// token from the client request
	token := httprouter.ParamsFromContext(r.Context()).ByName("token")

	// check if token is in served tokens
	service.mu.RLock()
	defer service.mu.RUnlock()

	if keyAuth, exists := service.tokens[token]; exists {
		service.logger.Debugf("writing challenge token to http-01 client: %s", token)

		w.Header().Set("Content-Type", "application/octet-stream")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(keyAuth))

		// done / exit
		return
	}

	// if token was not found, 404
	service.logger.Debugf("http-01 challenge token not found: %s", token)

	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("404 page not found"))

}
