package app

import (
	"net/http"
	"net/url"
)

var permittedHostnames = []string{"localhost", "127.0.0.1"}

func (app *Application) enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		allowedOrigin := "http://" + permittedHostnames[0] // default

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

		w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		w.Header().Set("Access-Control-Allow-Headers", "content-type, authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Methods", "PUT, DELETE")

		next.ServeHTTP(w, r)
	})
}
