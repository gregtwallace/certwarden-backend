package app

import (
	"context"
	"errors"
	"fmt"
	"legocerthub-backend/pkg/acme"
	"legocerthub-backend/pkg/challenges"
	"legocerthub-backend/pkg/domain/acme_accounts"
	"legocerthub-backend/pkg/domain/app/auth"
	"legocerthub-backend/pkg/domain/authorizations"
	"legocerthub-backend/pkg/domain/certificates"
	"legocerthub-backend/pkg/domain/download"
	"legocerthub-backend/pkg/domain/orders"
	"legocerthub-backend/pkg/domain/private_keys"
	"legocerthub-backend/pkg/httpclient"
	"legocerthub-backend/pkg/output"
	"legocerthub-backend/pkg/storage/sqlite"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"
)

// RunLeGoAPI starts the application
func RunLeGoAPI() {
	// app context for shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// create the app
	app, err := create(ctx)
	if err != nil {
		log.Panicf("failed to configure app: %s", err)
		return
	}
	defer app.CloseStorage()

	// configure webserver
	readTimeout := 5 * time.Second
	writeTimeout := 10 * time.Second
	// allow longer timeouts when in development
	if *app.config.DevMode {
		readTimeout = 15 * time.Second
		writeTimeout = 30 * time.Second
	}

	// http server config
	srv := &http.Server{
		Addr:         app.config.httpDomainAndPort(),
		Handler:      app.routes(),
		IdleTimeout:  1 * time.Minute,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
	}

	// var for redirect server (if needed)
	redirectSrv := &http.Server{}

	// configure and launch https if app succesfully got a cert
	if app.httpsCert != nil {
		// make tls config
		tlsConf, err := app.tlsConf()
		if err != nil {
			app.logger.Panicf("tls config problem: %s", err)
			return
		}

		// https server config
		srv.Addr = app.config.httpsDomainAndPort()
		srv.TLSConfig = tlsConf

		// configure and launch http redirect server
		redirectSrv = &http.Server{
			Addr: app.config.httpDomainAndPort(),
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Redirect(w, r, "https://"+app.config.httpsDomainAndPort()+r.RequestURI, http.StatusTemporaryRedirect)
			}),
			IdleTimeout:  1 * time.Minute,
			ReadTimeout:  readTimeout,
			WriteTimeout: writeTimeout,
		}
		app.logger.Infof("starting http redirect on %s", app.baseUrl())
		app.shutdownWaitgroup.Add(1)
		go func() {
			err := redirectSrv.ListenAndServe()
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				app.logger.Panic(err)
			}
			app.logger.Info("http redirect server shutdown complete")
			app.shutdownWaitgroup.Done()
		}()

		// launch https
		app.logger.Infof("starting lego-certhub (https) on %s", app.baseUrl())
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
		app.logger.Warnf("starting insecure lego-certhub (http) on %s", app.baseUrl())
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

	// disable shutdown context listener (allows for ctrl-c again to force close)
	stop()

	// shutdown the main web server (and redirect server)
	maxShutdownTime := 30 * time.Second
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), maxShutdownTime)
		defer cancel()

		err = srv.Shutdown(ctx)
		if err != nil {
			app.logger.Errorf("error shutting down http(s) server")
		}
	}()

	if app.httpsCert != nil {
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

	// close storage
	app.CloseStorage()

	// done
	app.logger.Info("shutdown complete")
	os.Exit(0)
}

// create creates an app object with logger, storage, and all needed
// services
func create(ctx context.Context) (*Application, error) {
	app := new(Application)
	var err error

	// parse config file
	cfg := readConfigFile()
	app.config = &cfg

	// logger (zap)
	app.initZapLogger()

	// shutdown context and waitgroup for graceful shutdown
	app.shutdownContext = ctx
	app.shutdownWaitgroup = new(sync.WaitGroup)

	// is the server in development mode?
	// this changes some basic things like: log level and connection timeouts
	// This does NOT prevent interactions with ACME production environment!
	if *app.config.DevMode {
		app.logger.Warn("development mode enabled (do not run in production)")
	}

	// create http client
	userAgent := fmt.Sprintf("LeGoCertHub/%s (%s; %s)", appVersion, runtime.GOOS, runtime.GOARCH)
	app.httpClient = httpclient.New(userAgent, *cfg.DevMode)

	// output service
	app.output, err = output.NewService(app)
	if err != nil {
		return nil, err
	}

	// acme services
	// use waitgroup to expedite directory fetching
	var wg sync.WaitGroup
	wgSize := 2

	wg.Add(wgSize)
	wgErrors := make(chan error, wgSize)

	// prod
	go func() {
		defer wg.Done()
		var err error
		app.acmeProd, err = acme.NewService(app, acmeProdUrl)
		wgErrors <- err
	}()

	// staging
	go func() {
		defer wg.Done()
		var err error
		app.acmeStaging, err = acme.NewService(app, acmeStagingUrl)
		wgErrors <- err
	}()

	wg.Wait()

	// check for errors
	close(wgErrors)
	for err = range wgErrors {
		if err != nil {
			return nil, err
		}
	}
	// end acme services

	// challenges
	app.challenges, err = challenges.NewService(app)
	if err != nil {
		return nil, err
	}

	// storage
	app.storage, err = sqlite.OpenStorage(app)
	if err != nil {
		return nil, err
	}

	// get app's tls cert
	// if fails, set to nil (will disable https)
	app.httpsCert, err = app.newAppCert()
	if err != nil {
		app.logger.Errorf("failed to configure https cert: %s", err)
		// if not https, and not dev mode, certain functions will be blocked
		if !*app.config.DevMode {
			app.logger.Error("certain functionality (e.g. pem downloads via API keys) will be disabled until the server is run in https mode")
		}
		app.httpsCert = nil
	}

	// users service
	app.auth, err = auth.NewService(app)
	if err != nil {
		return nil, err
	}

	// keys service
	app.keys, err = private_keys.NewService(app)
	if err != nil {
		return nil, err
	}

	// accounts service
	app.accounts, err = acme_accounts.NewService(app)
	if err != nil {
		return nil, err
	}

	// authorizations service
	app.authorizations, err = authorizations.NewService(app)
	if err != nil {
		return nil, err
	}

	// certificates service
	app.certificates, err = certificates.NewService(app)
	if err != nil {
		return nil, err
	}

	// orders service
	app.orders, err = orders.NewService(app)
	if err != nil {
		return nil, err
	}

	// download service
	app.download, err = download.NewService(app)
	if err != nil {
		return nil, err
	}

	return app, nil
}
