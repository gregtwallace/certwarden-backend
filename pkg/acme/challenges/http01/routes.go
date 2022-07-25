package http01

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// Routes creates the application's router and adds the routes. It also
// inserts the CORS middleware before returning the routes
func (service *Service) routes() http.Handler {
	router := httprouter.New()

	// acme challenge route, per rfc8555 8.3
	router.HandlerFunc(http.MethodGet, "/.well-known/acme-challenge/:token", service.challengeHandler)

	return enableCORS(router)
}
