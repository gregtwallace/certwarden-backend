package app

import (
	"legocerthub-backend/pkg/output"
	"net/http"
	"strings"

	"go.uber.org/zap"
)

// handler struct adds an error handling layer to handler functions. It includes
// logger so errors can be logged and the handlerFunc is in the custom handler
// format.
type handler struct {
	logger      *zap.SugaredLogger
	output      *output.Service
	handlerFunc customHandlerFunc
}

// customHandlerFunc is the form of the custom handler functions
type customHandlerFunc func(w http.ResponseWriter, r *http.Request) error

// handlerFunc converts a customer handeler function into a standard http.Handler
func (app *Application) makeHandler(handlerFunc customHandlerFunc) http.Handler {
	return handler{app.logger.SugaredLogger, app.output, handlerFunc}
}

// makeHandle creates a Handle on the app's router using the custom handler function
func (app *Application) makeHandle(method string, path string, handlerFunc customHandlerFunc) {
	app.router.Handler(method, path, app.makeHandler(handlerFunc))
}

// makeDownloadHandle is the same as makeHandle but adds some Info logging to keep track of
// clients accessing sensitive information
func (app *Application) makeDownloadHandle(method string, path string, handlerFunc customHandlerFunc) {
	downloadFunc := func(w http.ResponseWriter, r *http.Request) error {
		// don't directly log RequestURI since it can contain sensitive api keys in Legacy downloads
		// instead parse the URI to get just the non-sensitive portion
		// remove before download
		shortenedURI := r.RequestURI[strings.Index(r.RequestURI, "/download/")+1:]
		// split on /
		shortenedURIPieces := strings.SplitN(shortenedURI, "/", 4)
		// include first 3 pieces only
		shortenedURI = "/" + shortenedURIPieces[0] + "/" + shortenedURIPieces[1] + "/" + shortenedURIPieces[2]

		// log attempt
		app.logger.Infof("client %s attempting to download %s", r.RemoteAddr, shortenedURI)

		// handle attempt
		err := handlerFunc(w, r)
		if err != nil {
			app.logger.Infof("client %s error with download %s (%s)", r.RemoteAddr, shortenedURI, err)
			return err
		}

		// log success
		app.logger.Infof("client %s downloaded %s", r.RemoteAddr, shortenedURI)
		return nil
	}

	app.makeHandle(method, path, downloadFunc)
}

// ServeHTTP implements http.Handler. Essentially the handlerFunc is executed
// and the error is processed (logged and then written as JSON)
func (handler handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := handler.handlerFunc(w, r)

	// if there was an error, log it and write error JSON
	if err != nil {
		writeErr := handler.output.WriteErrorJSON(w, err)
		if writeErr != nil {
			handler.logger.Errorf("%s %s: failed to send error to client (failed to write error json: %s)", r.Method, r.URL.Path, writeErr)
		} else {
			handler.logger.Debugf("%s %s: error sent to client", r.Method, r.URL.Path)
		}
	}
	// else {
	// 	handler.logger.Debugf("%s %s: handled without error", r.Method, r.URL.Path)
	// }
}

// checkJwt is middleware that checks for a valid jwt
func (app *Application) checkJwt(next customHandlerFunc) customHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		_, err := app.auth.ValidAuthHeader(r.Header, w)
		if err != nil {
			return output.ErrUnauthorized
		}

		// if valid, execute next
		err = next(w, r)
		if err != nil {
			return err
		}

		return nil
	}
}

// makeSecureHandle is the same as makeHandle (makes a handle on the app's router) but it also
// adds the checkJwt middleware
func (app *Application) makeSecureHandle(method string, path string, handlerFunc customHandlerFunc) {
	app.makeHandle(method, path, app.checkJwt(handlerFunc))
}
