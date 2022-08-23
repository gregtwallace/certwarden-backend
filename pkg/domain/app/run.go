package app

import (
	"fmt"
	"legocerthub-backend/pkg/acme"
	"legocerthub-backend/pkg/domain/acme_accounts"
	"legocerthub-backend/pkg/domain/app/auth"
	"legocerthub-backend/pkg/domain/authorizations"
	"legocerthub-backend/pkg/domain/certificates"
	"legocerthub-backend/pkg/domain/orders"
	"legocerthub-backend/pkg/domain/private_keys"
	"legocerthub-backend/pkg/httpclient"
	"legocerthub-backend/pkg/output"
	"legocerthub-backend/pkg/storage/sqlite"
	"log"
	"net/http"
	"runtime"
	"sync"
	"time"
)

// RunLeGoAPI starts the application
func RunLeGoAPI() {
	// create the app
	app, err := create()
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
		Addr:         fmt.Sprintf("%s:%d", *app.config.Hostname, *app.config.HttpPort),
		Handler:      app.Routes(),
		IdleTimeout:  1 * time.Minute,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
	}

	// launch frontend if enabled
	if *app.config.Frontend.Enable {
		go app.runFrontend()
	}

	// configure and launch https if app succesfully got a cert
	if app.appCert != nil {
		// make tls config
		tlsConf, err := app.TlsConf()
		if err != nil {
			app.logger.Panicf("tls config problem: %s", err)
			return
		}

		// https server config
		srv.Addr = fmt.Sprintf("%s:%d", *app.config.Hostname, *app.config.HttpsPort)
		srv.TLSConfig = tlsConf

		// launch https
		app.logger.Infof("starting lego-certhub (https) on %s", srv.Addr)
		app.logger.Panic(srv.ListenAndServeTLS("", ""))
	} else {
		// if https failed, launch localhost only http server
		app.logger.Warnf("starting insecure lego-certhub (http) on %s", srv.Addr)
		app.logger.Panic(srv.ListenAndServe())
	}
}

// create creates an app object with logger, storage, and all needed
// services
func create() (*Application, error) {
	app := new(Application)
	var err error

	// parse config file
	cfg := readConfigFile()
	app.config = &cfg

	// logger (zap)
	app.initZapLogger()

	// is the server in development mode?
	// this changes some basic things like: log level and connection timeouts
	// This does NOT prevent interactions with ACME production environment!
	if *app.config.DevMode {
		app.logger.Error("key security measures disabled (development mode enabled)")
	}

	// create http client
	userAgent := fmt.Sprintf("LeGoCertHub/%s (%s; %s)", appVersion, runtime.GOOS, runtime.GOARCH)
	app.httpClient = httpclient.New(userAgent, *cfg.DevMode)

	// output service
	app.output, err = output.NewService(app)
	if err != nil {
		return nil, err
	}

	// storage
	app.storage, err = sqlite.OpenStorage()
	if err != nil {
		return nil, err
	}

	// get app's tls cert
	// if fails, set to nil (will disable https)
	app.appCert, err = app.newAppCert()
	if err != nil {
		app.logger.Errorf("failed to configure https cert: %s", err)
		// if not https, and not dev mode, certain functions will be blocked
		if !*app.config.DevMode {
			app.logger.Error("certain functionality (e.g. pem downloads via API keys) will be disabled until the server is run in https mode")
		}
		app.appCert = nil
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

	// acme services
	// use waitgroup to expedite directory fetching
	var wg sync.WaitGroup
	wgSize := 2

	wg.Add(wgSize)
	wgErrors := make(chan error, wgSize)

	// prod
	go func() {
		defer wg.Done()
		app.acmeProd, err = acme.NewService(app, acmeProdUrl)
		wgErrors <- err
	}()

	// staging
	go func() {
		defer wg.Done()
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

	return app, nil
}
