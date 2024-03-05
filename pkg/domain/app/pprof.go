package app

import (
	"context"
	"errors"
	"fmt"
	"legocerthub-backend/pkg/output"
	"net"
	"net/http"
	"net/http/pprof"
	"strings"

	"github.com/julienschmidt/httprouter"
)

// URL Paths for pprof
const pprofBasePath = ""
const pprofUrlPath = pprofBasePath + "/debug/pprof"

// pprofHandler handles all requests related to pprof
func pprofHandler(w http.ResponseWriter, r *http.Request) *output.Error {
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

	return nil
}

// pprofRouter creates the router for pprof by wrapping the pprofHandler
// in middleware
func (app *Application) makePprofRouter() http.Handler {
	router := httprouter.New()

	// Logger / handle custom handler func's error
	// do sensitive logging to show any access to pprof
	pprofHandlerFunc := middlewareApplyReturnValHandling(pprofHandler, true, app.logger.SugaredLogger, app.output)

	// actual handle path
	router.HandlerFunc(http.MethodGet, pprofUrlPath+"/*any", pprofHandlerFunc)

	// don't bother to apply HSTS here (main app access will handle it if desired)

	// security headers for browser
	// Note: CSP will complain and inline style blocking but who cares for pprof
	return middlewareApplyBrowserSecurityHeaders(router)
}

// startPprof starts the pprof http server on the configured port
func (app *Application) startPprof() error {
	// http server config
	srv := &http.Server{
		Addr:         app.config.pprofHttpServAddress(),
		Handler:      app.makePprofRouter(),
		IdleTimeout:  pprofServerIdleTimeout,
		ReadTimeout:  pprofServerReadTimeout,
		WriteTimeout: pprofServerWriteTimeout,
	}

	// if https, update config accordingly and log starting
	// note: http redirect server is not run for pprof

	if app.httpsCert != nil {
		// https server config
		srv.Addr = app.config.pprofHttpsServAddress()
		srv.TLSConfig = app.tlsConf()

		app.logger.Infof("starting pprof debugging (https) at %s", srv.Addr+pprofUrlPath)
	} else {
		app.logger.Infof("starting pprof debugging (http) at %s", srv.Addr+pprofUrlPath)
	}

	// create listener for web server
	ln, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		return fmt.Errorf("pprof server cannot bind to %s (%s)", srv.Addr, err)
	}

	// start server
	app.shutdownWaitgroup.Add(1)
	go func() {
		defer func() { _ = ln.Close }()
		defer app.shutdownWaitgroup.Done()

		// start server as https or http
		var err error
		if app.httpsCert != nil {
			err = srv.ServeTLS(ln, "", "")
		} else {
			err = srv.Serve(ln)
		}

		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			app.logger.Errorf("pprof server returned error (%s)", err)
		}
		app.logger.Info("pprof server shutdown complete")

	}()

	// shutdown server on shutdown signal
	go func() {
		<-app.shutdownContext.Done()

		ctx, cancel := context.WithTimeout(context.Background(), maxShutdownTime)
		defer cancel()

		err := srv.Shutdown(ctx)
		if err != nil {
			app.logger.Errorf("error shutting down pprof server")
		}
	}()

	return nil
}
