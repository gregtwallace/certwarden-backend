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
	// direct no caching, but include some backup options to try and cover all bases to ensure
	// the freshest response is always used
	w.Header().Set("Cache-Control", "no-store, no-cache, max-age=0, must-revalidate, proxy-revalidate")
	w.Header().Set("Pragma", "no-cache")
	// set valid but past date (again, to prevent caching)
	w.Header().Set("Expires", time.Time{}.Format(http.TimeFormat))

	// do not allow sniffing
	w.Header().Set("X-Content-Type-Options", "nosniff")

	// token from the client request
	token := httprouter.ParamsFromContext(r.Context()).ByName("token")

	// try to read resource
	keyAuth, exists := service.provisionedResources.Read(token)

	// resource not available, 404
	if !exists {
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
