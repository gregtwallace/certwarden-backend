package output

import (
	"net/http"
)

// WriteEmptyResponse writes an empty response to ResponseWriter with the specified
// status code.
func (service *Service) WriteEmptyResponse(w http.ResponseWriter, statusCode int) {
	// service.logger.Debugf("writing http status code to client (%s)", status)
	w.WriteHeader(statusCode)
}
