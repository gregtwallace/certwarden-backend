package frontend

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// routes creates the application's router and adds the routes. It also
// inserts the CORS middleware before returning the routes
func frontendRoutes() http.Handler {
	router := httprouter.New()

	// make file server handle for build dir
	router.ServeFiles("/*filepath", http.Dir(buildDir))

	return router
}
