package app

import (
	"net/http"
)

// enableHSTS applies the HTTP Strict Transport Security (HSTS) header to all
// responses from the server. If the server is not running in https mode, or if
// it has been disabled in config, the header is not applied.
func (app *Application) enableHSTS(next http.Handler) http.Handler {
	// do not send HSTS if running in http, or if it has been disabled explicitly
	if !app.IsHttps() || (app.config.DisableHSTS != nil && *app.config.DisableHSTS) {
		return next
	}

	// apply header and then proceed
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// set HSTS header
		w.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains")

		// serve next
		next.ServeHTTP(w, r)
	})
}
