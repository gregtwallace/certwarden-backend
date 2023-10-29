package app

import (
	"legocerthub-backend/pkg/domain/app/auth"
	"legocerthub-backend/pkg/output"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"go.uber.org/zap"
)

// handlerFunc is the type of the custom handler functions
type handlerFunc func(w http.ResponseWriter, r *http.Request) error

// router is the custom router implementation for app
type router struct {
	// services
	logger *zap.SugaredLogger
	output *output.Service
	auth   *auth.Service
	// actual router
	r *httprouter.Router
	// config options
	permittedCrossOrigins []string
}

// ServeHTTP calls the embedded router's ServeHTTP function
func (router *router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	router.r.ServeHTTP(w, r)
}

// handleAPIRouteInsecure creates a route on router intended for an INSECURE API route
func (router *router) handleAPIRouteInsecure(method string, path string, handlerFunc handlerFunc) {
	// CORS
	handlerFunc = middlewareApplyCORS(handlerFunc, router.permittedCrossOrigins)

	// Logger / handle custom handler func's error
	httpHandlerFunc := middlewareApplyErrorHandling(handlerFunc, false, router.logger, router.output)

	// make handler
	router.r.Handler(method, path, httpHandlerFunc)
}

// handleAPIRouteSecure creates a route on router intended for an authenticated API route
func (router *router) handleAPIRouteSecure(method string, path string, handlerFunc handlerFunc) {
	// JWT Auth
	handlerFunc = middlewareApplyAuthJWT(handlerFunc, router.auth)

	// CORS
	handlerFunc = middlewareApplyCORS(handlerFunc, router.permittedCrossOrigins)

	// Logger / handle custom handler func's error
	httpHandlerFunc := middlewareApplyErrorHandling(handlerFunc, false, router.logger, router.output)

	// make handler
	router.r.HandlerFunc(method, path, httpHandlerFunc)
}

// handleAPIRouteSecureDownload creates a route on router intended for downloading files via
// a logged in (SECURE) user.
func (router *router) handleAPIRouteSecureDownload(method string, path string, handlerFunc handlerFunc) {
	// Auth of API Keys is done by Downloads pkg, not here

	// CORS
	handlerFunc = middlewareApplyCORS(handlerFunc, router.permittedCrossOrigins)

	// Logger / handle custom handler func's error
	httpHandlerFunc := middlewareApplyErrorHandling(handlerFunc, true, router.logger, router.output)

	// make handler
	router.r.HandlerFunc(method, path, httpHandlerFunc)
}

// handleAPIRouteDownloadWithAPIKey creates a route on router intended for downloading files via
// their API key(s).
func (router *router) handleAPIRouteDownloadWithAPIKey(method string, path string, handlerFunc handlerFunc) {
	// Auth of API Keys is done by Downloads pkg, not here

	// NO CORS
	// downloads with api key should not cross-origin

	// Logger / handle custom handler func's error
	httpHandlerFunc := middlewareApplyErrorHandling(handlerFunc, true, router.logger, router.output)

	// make handler
	router.r.HandlerFunc(method, path, httpHandlerFunc)
}

// handleFrontend creates a route to serve content for the frontend
func (router *router) handleFrontend(method string, path string, handlerFunc handlerFunc) {
	// NO CORS
	// Frontend App should not cross-origin

	// Logger / handle custom handler func's error
	httpHandlerFunc := middlewareApplyErrorHandling(handlerFunc, true, router.logger, router.output)

	// make handler
	router.r.HandlerFunc(method, path, httpHandlerFunc)
}
