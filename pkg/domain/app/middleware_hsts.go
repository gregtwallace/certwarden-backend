package app

import "net/http"

// middlewareApplyHSTS applies the HTTP Strict Transport Security (HSTS) header.
// if the server is not https or if the header has been disabled in config, the header
// is not applied and this is a no-op.
func (app *Application) middlewareApplyHSTS(next handlerFunc) handlerFunc {
	// do not send HSTS if running in http, or if it has been disabled explicitly
	if !app.IsHttps() || (app.config.DisableHSTS != nil && *app.config.DisableHSTS) {
		return next
	}

	return func(w http.ResponseWriter, r *http.Request) error {
		// set HSTS header
		w.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains")

		// serve next
		return next(w, r)
	}
}
