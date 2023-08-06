package app

import (
	"net/http"
	"net/http/pprof"
	"strings"
)

// URL Paths for pprof
const pprofBasePath = baseUrlPath
const pprofUrlPath = pprofBasePath + "/debug/pprof"

// pprofHandler handles all requests related to pprof
func pprofHandler(w http.ResponseWriter, r *http.Request) error {
	// remove the URL base path
	r.URL.Path = strings.TrimPrefix(r.URL.Path, pprofBasePath)
	r.URL.RawPath = strings.TrimPrefix(r.URL.RawPath, pprofBasePath)

	// pprof route name to determine which pprof func to call
	pprofName, _ := strings.CutPrefix(r.URL.Path, "/debug/pprof/")

	// serve specific handlers (from pprof.go init(), otherwise default to Index)
	switch pprofName {
	case "cmdline":
		pprof.Cmdline(w, r)
	case "profile":
		pprof.Profile(w, r)
	case "symbol":
		pprof.Symbol(w, r)
	case "trace":
		pprof.Trace(w, r)
	default:
		// anything else, serve Index which also handles profiles
		pprof.Index(w, r)
	}

	// satisfy customHandlerFunc
	return nil
}
