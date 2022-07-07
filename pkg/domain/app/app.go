package app

import (
	"legocerthub-backend/pkg/acme"
	"legocerthub-backend/pkg/domain/acme_accounts"
	"legocerthub-backend/pkg/domain/private_keys"
	"legocerthub-backend/pkg/storage/sqlite"
	"log"

	"github.com/julienschmidt/httprouter"
)

const version = "0.0.1"

type Application struct {
	logger      *log.Logger
	router      *httprouter.Router
	storage     *sqlite.Storage
	keys        *private_keys.Service
	acmeProd    *acme.Service
	acmeStaging *acme.Service
	accounts    *acme_accounts.Service
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

func (app *Application) GetKeysService() *private_keys.Service {
	return app.keys
}

func (app *Application) GetAcmeProdService() *acme.Service {
	return app.acmeProd
}

func (app *Application) GetAcmeStagingService() *acme.Service {
	return app.acmeStaging
}
