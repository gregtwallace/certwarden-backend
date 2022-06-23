package app

import (
	"legocerthub-backend/pkg/acme"
	"legocerthub-backend/pkg/domain/acme_accounts"
	"legocerthub-backend/pkg/domain/private_keys"
	"legocerthub-backend/pkg/storage/sqlite"
	"log"
)

const version = "0.0.1"

type Application struct {
	logger      *log.Logger
	storage     *sqlite.Storage
	keys        *private_keys.Service
	acmeProd    *acme.Service
	acmeStaging *acme.Service
	accounts    *acme_accounts.Service
}

type appStatusDirectories struct {
	Production string `json:"prod"`
	Staging    string `json:"staging"`
}

type appStatus struct {
	Status          string               `json:"status"`
	Environment     string               `json:"environment"`
	Version         string               `json:"version"`
	AcmeDirectories appStatusDirectories `json:"acme_directories"`
}

/// return various needed components
func (app *Application) GetLogger() *log.Logger {
	return app.logger
}

// hacky workaround for storage since can't just combine into one interface
func (app *Application) GetKeyStorage() private_keys.Storage {
	return app.storage
}
func (app *Application) GetAccountStorage() acme_accounts.Storage {
	return app.storage
}

func (app *Application) GetAcmeProdService() *acme.Service {
	return app.acmeProd
}

func (app *Application) GetAcmeStagingService() *acme.Service {
	return app.acmeStaging
}
