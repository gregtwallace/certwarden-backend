package app

import (
	"legocerthub-backend/pkg/domain/acme_accounts"
	"legocerthub-backend/pkg/domain/private_keys"
	"legocerthub-backend/pkg/storage/sqlite"
	"legocerthub-backend/pkg/utils/acme_utils"
	"log"
)

const version = "0.0.1"

type Config struct {
	Host string
	Port int
	Env  string
}

type Application struct {
	Config         Config
	Logger         *log.Logger
	Storage        *sqlite.Storage
	AcmeProdDir    *acme_utils.AcmeDirectory
	AcmeStagingDir *acme_utils.AcmeDirectory
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
// hacky workaround for storage since can't just combine into one interface
func (app *Application) GetKeyStorage() private_keys.Storage {
	return app.Storage
}
func (app *Application) GetAccountStorage() acme_accounts.Storage {
	return app.Storage
}

func (app *Application) GetLogger() *log.Logger {
	return app.Logger
}

func (app *Application) GetProdDir() *acme_utils.AcmeDirectory {
	return app.AcmeProdDir
}

func (app *Application) GetStagingDir() *acme_utils.AcmeDirectory {
	return app.AcmeProdDir
}
