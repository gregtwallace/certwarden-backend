package http01internal

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// challengeHandler responds to the ACME http-01 challenge path. If the requested
// token exists in this service's tokens, the expected token content is sent back to
// the client. If the token is not in the service's tokens, a 404 reply is sent.
func (service *Service) challengeHandler(w http.ResponseWriter, r *http.Request) {
	var keyAuth string
	var exists bool
	var err error

	// token from the client request
	token := httprouter.ParamsFromContext(r.Context()).ByName("token")

	// lock tokens for reading
	service.mu.RLock()
	defer service.mu.RUnlock()

	// check token existence, if no token, write error 404 and return
	if keyAuth, exists = service.tokens[token]; !exists {
		service.logger.Debugf("http-01 challenge token not found: %s", token)

		w.WriteHeader(http.StatusNotFound)
		_, err = w.Write([]byte("404 page not found"))
		if err != nil {
			service.logger.Errorf("failed to write http-01 404 error: %s", err)
		}

		// done / exit
		return
	}

	// token was found, write it
	service.logger.Debugf("writing challenge token to http-01 client: %s", token)

	w.Header().Set("Content-Type", "application/octet-stream")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte(keyAuth))
	if err != nil {
		service.logger.Errorf("failed to write http-01 token: %s", err)
	}
}
