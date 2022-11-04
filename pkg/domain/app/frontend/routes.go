package frontend

import (
	"net/http"
	"os"
	"strings"

	"github.com/julienschmidt/httprouter"
)

// routes creates the application's router and adds the routes. It also
// inserts the CORS middleware before returning the routes
func (service *Service) frontendRoutes() http.Handler {
	router := httprouter.New()

	// file server for frontend
	fs := http.FileServer(http.Dir(buildDir))

	// no routes

	// handle all under not found (since no routes)
	router.NotFound = http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			// only potentially modify non '/' requests that dont contain a period in
			// the final segment (i.e. only modify requests for paths, specific file
			// names should still be able to 404)
			pathParts := strings.Split(r.URL.Path, "/")
			lastPart := pathParts[len(pathParts)-1]
			// check path is not / AND last part of the path does NOT contain a period (i.e. not a file)
			if r.URL.Path != "/" && !strings.Contains(lastPart, ".") {
				// check if request (as-is) exists
				fullPath := buildDir + r.URL.Path
				_, err := os.Stat(fullPath)
				if err != nil {
					// confirm error is file doesn't exist
					if !os.IsNotExist(err) {
						// if some other error, log it and return 404
						service.logger.Errorf("error serving frontend: %s", err)
						http.NotFound(w, r)
						return
					}

					// if doesn't exist, redirect to root (index)
					http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
					return
				}
			}

			// serve request from file server
			fs.ServeHTTP(w, r)
		})

	return router
}
