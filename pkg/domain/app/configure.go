package app

import (
	"fmt"
	"legocerthub-backend/pkg/acme"
	"legocerthub-backend/pkg/challenges/http01"
	"legocerthub-backend/pkg/domain/acme_accounts"
	"legocerthub-backend/pkg/domain/certificates"
	"legocerthub-backend/pkg/domain/private_keys"
	"legocerthub-backend/pkg/httpclient"
	"legocerthub-backend/pkg/output"
	"legocerthub-backend/pkg/storage/sqlite"
	"runtime"
	"sync"
)

// application version
const appVersion = "0.0.1"

// Directory URLs for Let's Encrypt
const acmeProdUrl string = "https://acme-v02.api.letsencrypt.org/directory"
const acmeStagingUrl string = "https://acme-staging-v02.api.letsencrypt.org/directory"

type Configuration struct {
	DevMode    bool
	Http01Port int
}

// CreateAndConfigure creates an app object with logger, storage, and all needed
// services
func CreateAndConfigure(config Configuration) (*Application, error) {
	app := new(Application)
	var err error

	// is the server in development mode?
	// this changes some basic things like: log level and connection timeouts
	// This does NOT prevent interactions with ACME production environment!
	app.devMode = config.DevMode

	// logger (zap)
	app.initZapLogger()

	// create http client
	userAgent := fmt.Sprintf("LeGoCertHub/%s (%s; %s)", appVersion, runtime.GOOS, runtime.GOARCH)
	app.httpClient = httpclient.New(userAgent, config.DevMode)

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
			break
		}
	}

	if err != nil {
		return nil, err
	}

	// accounts service
	app.accounts, err = acme_accounts.NewService(app)
	if err != nil {
		return nil, err
	}

	// http-01 challenge server
	app.http01, err = http01.NewService(app, config.Http01Port)
	if err != nil {
		return nil, err
	}

	// certificates service
	app.certificates, err = certificates.NewService(app)
	if err != nil {
		return nil, err
	}

	return app, nil
}

// CloseStorage closes the storage connection
func (app *Application) CloseStorage() {
	app.storage.Close()
}
