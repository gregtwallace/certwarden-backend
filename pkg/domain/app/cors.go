package app

import (
	"net/http"
)

// enableCORS applies CORS to an http.Handler and is intended to wrap the router
func (app *Application) enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// if additional CORS Permitted Origins are defined, check the origin against
		// them and add permissive header if match is found
		if len(app.config.CORSPermittedOrigins) > 0 {
			actualOrigin := r.Header.Get("Origin")
			for _, permittedOrigin := range app.config.CORSPermittedOrigins {
				// match found?
				if actualOrigin == permittedOrigin {
					w.Header().Set("Access-Control-Allow-Origin", actualOrigin)
					// once found, no need to check more
					break
				}
			}
		}

		// client to server headers
		// Access-Control-Allow-Origin not mandatory
		w.Header().Add("Access-Control-Allow-Headers", "authorization, content-type, x-no-retry")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Add("Access-Control-Allow-Methods", "DELETE, GET, HEAD, OPTIONS, POST, PUT")

		// server to client headers
		w.Header().Add("Access-Control-Expose-Headers", "content-disposition, content-type")

		next.ServeHTTP(w, r)
	})
}
