package app

import (
	"legocerthub-backend/pkg/output"
	"net/http"

	"go.uber.org/zap"
)

// handler struct adds an error handling layer to handler functions. It includes
// logger so errors can be logged and the handlerFunc is in the custom handler
// format.
type handler struct {
	logger      *zap.SugaredLogger
	handlerFunc customHandlerFunc
}

// customHandlerFunc is the form of the custom handler functions
type customHandlerFunc func(w http.ResponseWriter, r *http.Request) error

// handlerFunc converts a customer handeler function into a standard http.Handler
func (app *Application) makeHandler(handlerFunc customHandlerFunc) http.Handler {
	return handler{app.logger, handlerFunc}
}

// makeHandle creates a Handle on the app's router using the custom handler function
func (app *Application) makeHandle(method string, path string, handlerFunc customHandlerFunc) {
	app.router.Handler(method, path, app.makeHandler(handlerFunc))
}

// ServeHTTP implements http.Handler. Essentially the handlerFunc is executed
// and the error is processed (logged and then written as JSON)
func (handler handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := handler.handlerFunc(w, r)

	// if there was an error, log it and write error JSON
	if err != nil {
		writtenErr, writeErr := output.WriteErrorJSON(w, err)
		if writeErr != nil {
			handler.logger.Errorf("failed to send error to client for %s: failed to write json (%s)", r.URL.Path, writeErr)
		} else {
			handler.logger.Debugf("error sent to client for %s: %s", r.URL.Path, writtenErr)
		}
	} else {
		handler.logger.Debugf("%s %s: handled without error", r.Method, r.URL.Path)
	}
}
