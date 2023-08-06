package app

import (
	"net/http"
	_ "net/http/pprof"
	"strings"
)

// URL Paths for pprof
const pprofBasePath = baseUrlPath
const pprofUrlPath = pprofBasePath + "/debug/pprof"

// pprofHandler handles all requests related to pprof
func pprofHandler(w http.ResponseWriter, r *http.Request) error {
	// if at pprof root, ensure proper redirect (prevents '/pprof' (no trailing slash) from
	// redirecting incorrectly without base path)
	if r.URL.Path == pprofUrlPath {
		// add trailing slash
		http.Redirect(w, r, pprofUrlPath+"/", http.StatusPermanentRedirect)
	}

	// remove the URL root path
	r.URL.Path = strings.TrimPrefix(r.URL.Path, pprofBasePath)
	r.URL.RawPath = strings.TrimPrefix(r.URL.RawPath, pprofBasePath)

	// use default serve mix which pprof registers to
	http.DefaultServeMux.ServeHTTP(w, r)

	// satisfy customHandlerFunc
	return nil
}
