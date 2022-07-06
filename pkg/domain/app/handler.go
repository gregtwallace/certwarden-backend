package app

import (
	"legocerthub-backend/pkg/utils"
	"net/http"
)

// Handler struct adds an error handling layer to handler functions
type Handler struct {
	*Application
	Func func(w http.ResponseWriter, r *http.Request) error
}

// ServeHTTP implements http.Handler and will be the centralized point for error
// json writing
func (handler Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := handler.Func(w, r)

	// if there was an error, log it and write error JSON
	if err != nil {
		handler.Application.logger.Printf("Error %s: %s", r.URL.Path, err.Error())
		utils.WriteErrorJSON(w, err)
	}
}
