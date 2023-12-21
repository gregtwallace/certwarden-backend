package app

import (
	"context"
	"legocerthub-backend/pkg/challenges"
	"legocerthub-backend/pkg/datatypes"
	"legocerthub-backend/pkg/domain/acme_accounts"
	"legocerthub-backend/pkg/domain/acme_servers"
	"legocerthub-backend/pkg/domain/app/auth"
	"legocerthub-backend/pkg/domain/app/backup"
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
	"sync"
	"time"

	"go.uber.org/zap"
)

// application version
const appVersion = "0.17.0"

// config version
// increment any time there is a breaking change between versions
const appConfigVersion = 3

// data storage root
const dataStorageRootPath = "./data"
const dataStorageAppDataDirName = "app"
const dataStorageAppDataPath = dataStorageRootPath + "/" + dataStorageAppDataDirName

// http server timeouts
const httpServerReadTimeout = 5 * time.Second
const httpServerWriteTimeout = 10 * time.Second
const httpServerIdleTimeout = 1 * time.Minute

const pprofServerReadTimeout = httpServerReadTimeout
const pprofServerWriteTimeout = 30 * time.Second
const pprofServerIdleTimeout = httpServerIdleTimeout

// appLogger is a SugaredLogger + a close function to sync (flush) the
// logger and to close the underlying file
type appLogger struct {
	*zap.SugaredLogger
	syncAndClose func()
}

// Application is the main app struct
type Application struct {
	restart           bool
	config            *config
	logger            *appLogger
	output            *output.Service
	backup            *backup.Service
	shutdownContext   context.Context
	shutdown          func(restart bool)
	shutdownWaitgroup *sync.WaitGroup
	httpsCert         *datatypes.SafeCert
	httpClient        *httpclient.Client
	router            http.Handler
	storage           *sqlite.Storage
	acmeServers       *acme_servers.Service
	challenges        *challenges.Service
	updater           *updater.Service
	auth              *auth.Service
	keys              *private_keys.Service
	accounts          *acme_accounts.Service
	authorizations    *authorizations.Service
	orders            *orders.Service
	certificates      *certificates.Service
	download          *download.Service
}

// return various app parts which are used as needed by services
func (app *Application) GetAppVersion() string {
	return appVersion
}

func (app *Application) GetDataStorageRootPath() string {
	return dataStorageRootPath
}

func (app *Application) GetDataStorageAppDataPath() string {
	return dataStorageAppDataPath
}

// LockSQLForBackup locks sql storage from writes so that a copy can be read without
// the risk of corruption. It returns a function to unlock the db after the backup
// is completed.
func (app *Application) LockSQLForBackup() (unlockFunc func(), err error) {
	// if storage hasn't been setup yet, no risk of writes so this is no-op
	if app.storage != nil {
		return app.storage.LockDBForBackup()
	}

	// no-up unlock if there wasn't a real lock
	return func() {}, nil
}

func (app *Application) CreateBackupOnDisk() error {
	_, err := app.backup.CreateBackupOnDisk()
	return err
}

func (app *Application) GetConfigVersion() int {
	return appConfigVersion
}

func (app *Application) GetLogger() *zap.SugaredLogger {
	return app.logger.SugaredLogger
}

// is the server running https or not?
func (app *Application) IsHttps() bool {
	return app.httpsCert != nil
}

// are any cross origins allowed?
func (app *Application) AllowsSomeCrossOrigin() bool {
	return len(app.config.CORSPermittedCrossOrigins) > 0
}

func (app *Application) GetHttpClient() *httpclient.Client {
	return app.httpClient
}

func (app *Application) GetOutputter() *output.Service {
	return app.output
}

func (app *Application) GetChallengesService() *challenges.Service {
	return app.challenges
}

// hacky workaround for storage since can't just combine into one interface
func (app *Application) GetAuthStorage() auth.Storage {
	return app.storage
}
func (app *Application) GetKeyStorage() private_keys.Storage {
	return app.storage
}
func (app *Application) GetAcmeServerStorage() acme_servers.Storage {
	return app.storage
}
func (app *Application) GetAccountStorage() acme_accounts.Storage {
	return app.storage
}
func (app *Application) GetCertificatesStorage() certificates.Storage {
	return app.storage
}
func (app *Application) GetOrderStorage() orders.Storage {
	return app.storage
}
func (app *Application) GetDownloadStorage() download.Storage {
	return app.storage
}

//

func (app *Application) GetKeysService() *private_keys.Service {
	return app.keys
}

func (app *Application) GetAcmeServerService() *acme_servers.Service {
	return app.acmeServers
}

func (app *Application) GetAcctsService() *acme_accounts.Service {
	return app.accounts
}

func (app *Application) GetAuthsService() *authorizations.Service {
	return app.authorizations
}

func (app *Application) GetCertificatesService() *certificates.Service {
	return app.certificates
}

// shutdown related
func (app *Application) GetShutdownContext() context.Context {
	return app.shutdownContext
}

func (app *Application) GetShutdownWaitGroup() *sync.WaitGroup {
	return app.shutdownWaitgroup
}
