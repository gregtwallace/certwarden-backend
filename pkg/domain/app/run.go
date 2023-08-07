package app

import (
	"context"
	"errors"
	"fmt"
	"legocerthub-backend/pkg/challenges"
	"legocerthub-backend/pkg/domain/acme_accounts"
	"legocerthub-backend/pkg/domain/acme_servers"
	"legocerthub-backend/pkg/domain/app/auth"
	"legocerthub-backend/pkg/domain/app/updater"
	"legocerthub-backend/pkg/domain/authorizations"
	"legocerthub-backend/pkg/domain/certificates"
	"legocerthub-backend/pkg/domain/download"
	"legocerthub-backend/pkg/domain/orders"
	"legocerthub-backend/pkg/domain/private_keys"
	"legocerthub-backend/pkg/httpclient"
	"legocerthub-backend/pkg/output"
	"legocerthub-backend/pkg/storage/sqlite"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

const maxShutdownTime = 30 * time.Second

// RunLeGoAPI starts the application
func RunLeGoAPI() {
	// app context for shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// create the app
	app, err := create(ctx)
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
		// make tls config
		tlsConf, err := app.tlsConf()
		if err != nil {
			app.logger.Panicf("tls config problem: %s", err)
			return
		}

		// https server config
		srv.Addr = app.config.httpsServAddress()
		srv.TLSConfig = tlsConf

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

	// disable shutdown context listener (allows for ctrl-c again to force close)
	stop()

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

	// logger pre-config reading
	app.initZapLogger()

	// startup log
	app.logger.Infof("starting LeGo CertHub v%s", appVersion)

	// make data dir if doesn't exist
	_, err = os.Stat(dataStoragePath)
	if errors.Is(err, os.ErrNotExist) {
		// create data dir
		err = os.Mkdir(dataStoragePath, 0700)
		if err != nil {
			app.logger.Errorf("failed to make data storage directory (%s)", err)
			return app, err
		}
	}

	// parse config file
	err = app.readConfigFile()
	if err != nil {
		app.logger.Errorf("failed to read app config file (%s)", err)
		return app, err
	}

	// logger (re-init using config settings)
	app.initZapLogger()

	// config file version check
	if app.config.ConfigVersion != configVersion {
		app.logger.Errorf("config.yaml config_version (%d) does not match app (%d), review config change log", app.config.ConfigVersion, configVersion)
	}

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
	app.httpClient = httpclient.New(userAgent, *app.config.DevMode)

	// output service
	app.output, err = output.NewService(app)
	if err != nil {
		app.logger.Errorf("failed to configure app output (%s)", err)
		return app, err
	}

	// storage
	app.storage, err = sqlite.OpenStorage(app, dataStoragePath)
	if err != nil {
		app.logger.Errorf("failed to configure app storage (%s)", err)
		return app, err
	}

	// acmeServers
	app.acmeServers, err = acme_servers.NewService(app)
	if err != nil {
		app.logger.Errorf("failed to configure app acme servers (%s)", err)
		return app, err
	}

	// challenges
	app.challenges, err = challenges.NewService(app, &app.config.Challenges)
	if err != nil {
		app.logger.Errorf("failed to configure app challenges (%s)", err)
		return app, err
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

	// app updater service
	app.updater, err = updater.NewService(app, &app.config.Updater)
	if app.updater == nil || err != nil {
		app.logger.Errorf("failed to configure app updater (%s)", err)
		return app, err
	}

	// users service
	app.auth, err = auth.NewService(app)
	if err != nil {
		app.logger.Errorf("failed to configure app authentication (%s)", err)
		return app, err
	}

	// keys service
	app.keys, err = private_keys.NewService(app)
	if err != nil {
		app.logger.Errorf("failed to configure app keys (%s)", err)
		return app, err
	}

	// accounts service
	app.accounts, err = acme_accounts.NewService(app)
	if err != nil {
		app.logger.Errorf("failed to configure app accounts (%s)", err)
		return app, err
	}

	// authorizations service
	app.authorizations, err = authorizations.NewService(app)
	if err != nil {
		app.logger.Errorf("failed to configure app authorizations (%s)", err)
		return app, err
	}

	// certificates service
	app.certificates, err = certificates.NewService(app)
	if err != nil {
		app.logger.Errorf("failed to configure app certificates (%s)", err)
		return app, err
	}

	// orders service
	app.orders, err = orders.NewService(app, &app.config.Orders)
	if err != nil {
		app.logger.Errorf("failed to configure app orders (%s)", err)
		return app, err
	}

	// download service
	app.download, err = download.NewService(app)
	if err != nil {
		app.logger.Errorf("failed to configure app download (%s)", err)
		return app, err
	}

	return app, nil
}
