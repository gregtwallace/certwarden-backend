package app

import (
	"net/http"
)

// middlewareApplyHSTS applies the HTTP Strict Transport Security (HSTS) header.
func middlewareApplyHSTS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// set HSTS header
		w.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains")

		// serve next
		next.ServeHTTP(w, r)
	})
}

// middlewareApplyReferrerPolicy applies the Referrer-Policy header. LeGo does
// not currently use the Referer header for any reason, so the strictest policy
// 'no-referrer' is applied
func middlewareApplyReferrerPolicy(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// set HSTS header
		w.Header().Set("Referrer-Policy", "no-referrer")

		// serve next
		next.ServeHTTP(w, r)
	})
}
