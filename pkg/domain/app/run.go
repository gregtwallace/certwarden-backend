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
		app.logger.Panicf("panic: failed to configure app: %s", err)
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

	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", *app.config.Hostname, *app.config.Port),
		Handler:      app.Routes(),
		IdleTimeout:  1 * time.Minute,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
	}

	// launch webserver
	app.GetLogger().Infof("starting lego-certhub on %s:%d", *app.config.Hostname, *app.config.Port)
	err = srv.ListenAndServe()
	if err != nil {
		app.GetLogger().Panicf("panic: failed to start http server: %s", err)
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
		app.logger.Warn("Development mode ENABLED. Key security measures DISABLED.")
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
