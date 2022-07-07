package app

import (
	"legocerthub-backend/pkg/output"
	"log"
	"net/http"
)

// Handler struct adds an error handling layer to handler functions. It includes
// logger so errors can be logged and the handlerFunc is in the format the handler
// functions will be in
type Handler struct {
	logger      *log.Logger
	handlerFunc func(w http.ResponseWriter, r *http.Request) error
}

// Handler creates a Handler on the app's router using the custom Handler struct
func (app *Application) Handler(method string, path string, handler func(http.ResponseWriter, *http.Request) error) {
	app.router.Handler(method, path, Handler{app.logger, handler})
}

// ServeHTTP implements http.Handler. Essentially the handlerFunc is executed
// and the error is processed (logged and then written as JSON)
func (handler Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := handler.handlerFunc(w, r)

	// if there was an error, log it and write error JSON
	if err != nil {
		handler.logger.Printf("error %s: %s", r.URL.Path, err.Error())

		writeErr := output.WriteErrorJSON(w, err)
		if writeErr != nil {
			handler.logger.Printf("error %s: failed to write json (%s)", r.URL.Path, writeErr.Error())
		}
	}
}
