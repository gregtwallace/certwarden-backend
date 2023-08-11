package app

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const maxShutdownTime = 30 * time.Second

// RunLeGoAPI starts the application and also contains restart logic
func RunLeGoAPI() {
	// run and restart if appropriate
	for {
		restart := run()

		if !restart {
			break
		}
	}

	// done, app exit
}

// run starts an instance of the application
func run() (restart bool) {
	// create the app
	app, err := create()
	if err != nil {
		app.logger.Errorf("failed to create app (%s)", err)
		os.Exit(1)
		return
	}
	defer app.CloseStorage()

	// start pprof if enabled or in dev mode
	if *app.config.DevMode || *app.config.EnablePprof {
		app.startPprof()
	}

	// configure webserver
	readTimeout := 5 * time.Second
	writeTimeout := 10 * time.Second
	// allow longer timeouts when in development
	if *app.config.DevMode {
		readTimeout = 15 * time.Second
		writeTimeout = 120 * time.Second
	}

	// http server config
	srv := &http.Server{
		Addr:         app.config.httpServAddress(),
		Handler:      app.routes(),
		IdleTimeout:  1 * time.Minute,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
	}

	// var for redirect server (if needed)
	redirectSrv := &http.Server{}

	// configure and launch https if app succesfully got a cert
	if app.httpsCert != nil {
		// https server config
		srv.Addr = app.config.httpsServAddress()
		srv.TLSConfig = app.tlsConf()

		// configure and launch http redirect server
		if *app.config.EnableHttpRedirect {
			redirectSrv = &http.Server{
				Addr: app.config.httpServAddress(),
				Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					// remove port (if present) to get request hostname alone (since changing port)
					hostName, _, _ := strings.Cut(r.Host, ":")

					// build redirect address
					newAddr := "https://" + hostName + ":" + strconv.Itoa(*app.config.HttpsPort) + r.RequestURI

					http.Redirect(w, r, newAddr, http.StatusTemporaryRedirect)
				}),
				IdleTimeout:  1 * time.Minute,
				ReadTimeout:  readTimeout,
				WriteTimeout: writeTimeout,
			}
			app.logger.Infof("starting http redirect bound to %s", app.config.httpServAddress())
			app.shutdownWaitgroup.Add(1)
			go func() {
				err := redirectSrv.ListenAndServe()
				if err != nil && !errors.Is(err, http.ErrServerClosed) {
					app.logger.Panic(err)
				}
				app.logger.Info("http redirect server shutdown complete")
				app.shutdownWaitgroup.Done()
			}()
		}

		// launch https
		app.logger.Infof("starting lego-certhub (https) bound to %s", app.config.httpsServAddress())
		app.shutdownWaitgroup.Add(1)
		go func() {
			err := srv.ListenAndServeTLS("", "")
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				app.logger.Panic(err)
			}
			app.logger.Info("https server shutdown complete")
			app.shutdownWaitgroup.Done()
		}()

	} else {
		// if https failed, launch localhost only http server
		app.logger.Warnf("starting insecure lego-certhub (http) bound to %s", app.config.httpServAddress())
		app.shutdownWaitgroup.Add(1)
		go func() {
			err := srv.ListenAndServe()
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				app.logger.Panic(err)
			}
			app.logger.Info("http server shutdown complete")
			app.shutdownWaitgroup.Done()
		}()
	}

	// shutdown logic
	// wait for shutdown context to signal
	<-app.shutdownContext.Done()

	// shutdown the main web server (and redirect server)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), maxShutdownTime)
		defer cancel()

		err = srv.Shutdown(ctx)
		if err != nil {
			app.logger.Errorf("error shutting down http(s) server")
		}
	}()

	if app.httpsCert != nil && *app.config.EnableHttpRedirect {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), maxShutdownTime)
			defer cancel()

			err = redirectSrv.Shutdown(ctx)
			if err != nil {
				app.logger.Errorf("error shutting down http redirect server")
			}
		}()
	}

	// wait for each component/service to shutdown
	// but also implement a maxWait chan to force close (panic)
	maxWait := 2 * time.Minute
	waitChan := make(chan struct{})

	// close wait chan when wg finishes waiting
	go func() {
		defer close(waitChan)
		app.shutdownWaitgroup.Wait()
	}()

	select {
	case <-waitChan:
		// continue, normal
	case <-time.After(maxWait):
		// timed out
		app.logger.Panic("shutdown procedure failed (timed out)")
	}

	// log based on if trying to restart
	if app.restart {
		app.logger.Infow("shutdown complete, restarting lego")
	} else {
		app.logger.Info("shutdown complete")
	}

	// restart return var
	restart = app.restart

	// nil app
	app = nil

	return restart
}
