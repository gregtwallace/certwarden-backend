package app

import "net/http"

// middlewareApplyReferrerPolicy applies the Referrer-Policy header.
func middlewareApplyReferrerPolicy(next handlerFunc) handlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		// set HSTS header
		w.Header().Set("Referrer-Policy", "no-referrer")

		// serve next
		return next(w, r)
	}
}
