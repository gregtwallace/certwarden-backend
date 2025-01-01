package app

import (
	"certwarden-backend/pkg/challenges"
	"certwarden-backend/pkg/datatypes/safecert"
	"certwarden-backend/pkg/domain/acme_accounts"
	"certwarden-backend/pkg/domain/acme_servers"
	"certwarden-backend/pkg/domain/app/auth"
	"certwarden-backend/pkg/domain/app/backup"
	"certwarden-backend/pkg/domain/app/updater"
	"certwarden-backend/pkg/domain/authorizations"
	"certwarden-backend/pkg/domain/certificates"
	"certwarden-backend/pkg/domain/download"
	"certwarden-backend/pkg/domain/orders"
	"certwarden-backend/pkg/domain/private_keys"
	"certwarden-backend/pkg/httpclient"
	"certwarden-backend/pkg/output"
	"certwarden-backend/pkg/storage/sqlite"
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
)

// create creates an app object with logger, storage, and all needed
// services
func create() (*Application, error) {
	app := new(Application)
	var err error

	// logger pre-config reading
	app.initZapLogger()

	// startup log
	app.logger.Infof("starting Cert Warden v%s", appVersion)

	// output service
	app.output, err = output.NewService(app)
	if err != nil {
		app.logger.Errorf("failed to configure app output (%s)", err)
		return app, err
	}

	// app backup service
	app.backup, err = backup.NewService(app)
	if app.backup == nil || err != nil {
		app.logger.Errorf("failed to configure app backup (%s)", err)
		return app, err
	}

	// make app data dir if doesn't exist
	_, err = os.Stat(dataStorageAppDataPath)
	if errors.Is(err, os.ErrNotExist) {
		// create data dir
		err = os.Mkdir(dataStorageAppDataPath, 0700)
		if err != nil {
			app.logger.Errorf("failed to make data storage directory (%s)", err)
			return app, err
		}
	}

	// parse config file (also create if doesn't exist)
	err = app.loadConfigFile()
	if err != nil {
		app.logger.Errorf("failed to read app config file (%s)", err)
		return app, err
	}

	// logger (re-init using config settings)
	app.initZapLogger()

	// config file version check
	if *app.config.ConfigVersion != appConfigVersion {
		app.logger.Errorf("config.yaml config_version (%d) does not match app (%d), review config change log", *app.config.ConfigVersion, appConfigVersion)
	}

	// context for shutdown OS signal
	osSignalCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	// wait for the OS signal and then stop listening and call shutdown
	go func() {
		<-osSignalCtx.Done()

		// disable shutdown context listener (allows for ctrl-c again to force close)
		stop()

		// log os signal call unless shutdown was already triggered somewhere else
		select {
		case <-app.shutdownContext.Done():
			// no-op
		default:
			app.logger.Info("os signal received for shutdown")
		}

		// do shutdown
		app.shutdown(false)
	}()

	// shutdown context with func to call graceful shutdown
	shutdownContext, doShutdown := context.WithCancel(context.Background())
	app.shutdownContext = shutdownContext
	// this func is called to shutdown app, only run once to avoid overwrite of
	// restart value
	once := &sync.Once{}
	app.shutdown = func(restart bool) {
		once.Do(func() {
			if restart {
				app.logger.Info("graceful restart triggered")
			} else {
				app.logger.Info("graceful shutdown triggered")
			}

			app.restart = restart

			// stop os signal listening
			stop()

			// shutdown
			doShutdown()
		})
	}

	// wait group for graceful shutdown
	app.shutdownWaitgroup = new(sync.WaitGroup)

	// create http client
	userAgent := fmt.Sprintf("CertWarden/%s (%s; %s)", appVersion, runtime.GOOS, runtime.GOARCH)
	app.httpClient = httpclient.New(userAgent)

	// start automatic backup service
	app.backup.StartAutoBackupService(app, &app.config.Backup)

	// if db file does not exist at new location, check old location and move file
	// from old to new (if exists at old location)
	if _, err := os.Stat(dataStorageAppDataPath + "/" + sqlite.DbFilename); errors.Is(err, os.ErrNotExist) {
		// stat old location
		if _, err := os.Stat(dataStorageRootPath + "/" + sqlite.DbFilename); err == nil {
			// exists at old location, move it
			err = os.Rename(dataStorageRootPath+"/"+sqlite.DbFilename, dataStorageAppDataPath+"/"+sqlite.DbFilename)
			if err != nil {
				app.logger.Errorf("failed to move app db file from old location to new location (%s)", err)
				return app, err
			}
			app.logger.Infof("storage database moved from %s to %s", dataStorageRootPath+"/"+sqlite.DbFilename, dataStorageAppDataPath+"/"+sqlite.DbFilename)
		}
	}

	// storage
	app.storage, err = sqlite.OpenStorage(app)
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

	// load app's tls cert
	// if error, server will instead operate over http
	app.httpsCert = safecert.NewSafeCert(app.httpClient, app.shutdownWaitgroup, shutdownContext)
	err = app.LoadHttpsCertificate()
	if err != nil {
		// failed = not https
		app.logger.Errorf("failed to configure https cert: %s", err)
		app.httpsCert = nil
	}

	// app updater service
	app.updater, err = updater.NewService(app, &app.config.Updater)
	if app.updater == nil || err != nil {
		app.logger.Errorf("failed to configure app updater (%s)", err)
		return app, err
	}

	// users service
	app.auth, err = auth.NewService(app, &app.config.Auth)
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

	// make router
	app.makeRouterAndRoutes()

	return app, nil
}
