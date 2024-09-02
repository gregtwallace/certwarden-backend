package http01internal

import (
	"bytes"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
)

// challengeHandler responds to the ACME http-01 challenge path. If the requested
// token exists in this service's resources, the keyAuth bytes are sent back to
// the client. If the token is not in the service's resources, a 404 reply is sent.
func (service *Service) challengeHandler(w http.ResponseWriter, r *http.Request) {
	// token from the client request
	token := httprouter.ParamsFromContext(r.Context()).ByName("token")

	// try to read resource
	keyAuth, err := service.provisionedResources.Read(token)

	// resource not available, 404
	if err != nil {
		service.logger.Debugf("http-01 challenge resource %s not found", token)

		// write status 404
		w.WriteHeader(http.StatusNotFound)

		// done / exit
		return
	}

	// token was found, write it
	service.logger.Debugf("writing resource (name: %s) to http-01 client", token)

	// convert value to content reader for output
	contentReader := bytes.NewReader([]byte(keyAuth))

	// Set Content-Type explicitly
	w.Header().Set("Content-Type", "application/octet-stream")

	// ServeContent (filename is not needed here since Content-Type is set explicitly above)
	http.ServeContent(w, r, "", time.Time{}, contentReader)
}
