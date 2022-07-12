package app

import (
	"legocerthub-backend/pkg/acme"
	"legocerthub-backend/pkg/domain/acme_accounts"
	"legocerthub-backend/pkg/domain/private_keys"
	"legocerthub-backend/pkg/storage/sqlite"
)

// Directory URLs for Let's Encrypt
const acmeProdUrl string = "https://acme-v02.api.letsencrypt.org/directory"
const acmeStagingUrl string = "https://acme-staging-v02.api.letsencrypt.org/directory"

// CreateAndConfigure creates an app object with logger, storage, and all needed
// services
func CreateAndConfigure(devMode bool) (*Application, error) {
	app := new(Application)
	var err error

	// logger (zap)
	app.initZapLogger()

	// is the server in development mode?
	// this changes some basic things like: log level, connection timeouts, and
	// json indent
	// This does NOT prevent interactions with ACME production environment!
	// TODO: Implement these changes
	app.devMode = devMode

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
	app.acmeProd, err = acme.NewService(app, acmeProdUrl)
	if err != nil {
		return nil, err
	}

	app.acmeStaging, err = acme.NewService(app, acmeStagingUrl)
	if err != nil {
		return nil, err
	}

	// accounts service
	app.accounts, err = acme_accounts.NewService(app)
	if err != nil {
		return nil, err
	}

	return app, nil
}

// CloseStorage closes the storage connection
func (app *Application) CloseStorage() {
	app.storage.Close()
}
