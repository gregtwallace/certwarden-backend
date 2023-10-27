package app

import (
	"net/http"
)

// handlerFunc is the type of the custom handler functions
type handlerFunc func(w http.ResponseWriter, r *http.Request) error

// handleAPIRouteInsecure creates a route on router intended for an INSECURE API route
func (app *Application) handleAPIRouteInsecure(method string, path string, handlerFunc handlerFunc) {
	// HSTS
	handlerFunc = app.middlewareApplyHSTS(handlerFunc)

	// Logger
	httpHandlerFunc := app.middlewareApplyLogger(handlerFunc, false)

	// CORS
	httpHandlerFunc = app.middlewareApplyCORS(httpHandlerFunc)

	// make handler
	app.router.Handler(method, path, httpHandlerFunc)
}

// handleAPIRouteSecure creates a route on router intended for an authenticated API route
func (app *Application) handleAPIRouteSecure(method string, path string, handlerFunc handlerFunc) {
	// JWT Auth
	handlerFunc = app.middlewareApplyAuthJWT(handlerFunc)

	// HSTS
	handlerFunc = app.middlewareApplyHSTS(handlerFunc)

	// Logger (handlerFunc -> http.HandlerFunc)
	httpHandlerFunc := app.middlewareApplyLogger(handlerFunc, false)

	// CORS
	httpHandlerFunc = app.middlewareApplyCORS(httpHandlerFunc)

	// make handler
	app.router.HandlerFunc(method, path, httpHandlerFunc)
}

// handleAPIRouteSecureDownload creates a route on router intended for downloading files via
// a logged in (SECURE) user.
func (app *Application) handleAPIRouteSecureDownload(method string, path string, handlerFunc handlerFunc) {
	// Auth of API Keys is done by Downloads pkg, not here

	// HSTS
	handlerFunc = app.middlewareApplyHSTS(handlerFunc)

	// Logger (handlerFunc -> http.HandlerFunc)
	httpHandlerFunc := app.middlewareApplyLogger(handlerFunc, true)

	// CORS
	httpHandlerFunc = app.middlewareApplyCORS(httpHandlerFunc)

	// make handler
	app.router.HandlerFunc(method, path, httpHandlerFunc)
}

// handleAPIRouteDownloadWithAPIKey creates a route on router intended for downloading files via
// their API key(s).
func (app *Application) handleAPIRouteDownloadWithAPIKey(method string, path string, handlerFunc handlerFunc) {
	// Auth of API Keys is done by Downloads pkg, not here

	// HSTS
	handlerFunc = app.middlewareApplyHSTS(handlerFunc)

	// Logger (handlerFunc -> http.HandlerFunc)
	httpHandlerFunc := app.middlewareApplyLogger(handlerFunc, true)

	// NO CORS
	// downloads with api key should not cross-origin

	// make handler
	app.router.HandlerFunc(method, path, httpHandlerFunc)
}

// handleFrontend creates a route to serve content for the frontend
func (app *Application) handleFrontend(method string, path string, handlerFunc handlerFunc) {
	// no auth to load the frontend app

	// HSTS
	handlerFunc = app.middlewareApplyHSTS(handlerFunc)

	// Logger (handlerFunc -> http.HandlerFunc)
	httpHandlerFunc := app.middlewareApplyLogger(handlerFunc, true)

	// NO CORS
	// Frontend App should not cross-origin

	// make handler
	app.router.HandlerFunc(method, path, httpHandlerFunc)
}
