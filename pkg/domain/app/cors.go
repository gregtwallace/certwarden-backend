package app

import (
	"net/http"
	"net/url"
)

// by default, these are always allowed
var defaultHostnames = []string{"localhost", "127.0.0.1"}

func (app *Application) enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// generic default origin if origin isn't found in list (will be rejected)
		allowedOrigin := "https://" + defaultHostnames[0]

		// add config hostname to approved list
		permittedHostnames := []string{*app.config.Hostname}
		permittedHostnames = append(permittedHostnames, defaultHostnames...)

		// allow any scheme and/or port from a permitted origin
		url, err := url.ParseRequestURI(r.Header.Get("Origin"))
		if err == nil {
			for _, hostname := range permittedHostnames {
				if hostname == url.Hostname() {
					allowedOrigin = url.String()
					break
				}
			}
		}

		// client to server headers
		w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		w.Header().Add("Access-Control-Allow-Headers", "authorization, content-type, x-no-retry")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Add("Access-Control-Allow-Methods", "PUT, DELETE")

		// server to client headers
		w.Header().Add("Access-Control-Expose-Headers", "content-disposition, content-type")

		next.ServeHTTP(w, r)
	})
}
