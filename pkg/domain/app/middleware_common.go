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

// middlewareApplyBrowserSecurityHeaders applies a number of security headers to
// reduce the danger of things like click jacking or loading malicious data. these
// are under common as additional precaution. even though many routes are not
// intended for use in a browser, these don't hurt.
func middlewareApplyBrowserSecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1) header: Content-Security-Policy - disable loading of other content
		// Note: this can be overwritten by the frontend handler, when needed
		var contentSecurityPolicy = []string{
			"default-src 'none'",
			"base-uri 'none'",
			"form-action 'none'",
			"frame-ancestors 'none'",
		}

		csp := ""
		for _, s := range contentSecurityPolicy {
			csp += s + "; "
		}

		w.Header().Set("Content-Security-Policy", csp)

		// 2) header: X-Content-Type-Options - no MIME type sniffing
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// 3) header: X-Frame-Options - do NOT allow frames
		w.Header().Set("X-Frame-Options", "deny")

		// 4) header: Referrer - tell browser to never send Referer header
		w.Header().Set("Referrer-Policy", "no-referrer")

		// serve next
		next.ServeHTTP(w, r)
	})
}
