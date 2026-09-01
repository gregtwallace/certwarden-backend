package storage

import (
	"certwarden-backend/pkg/storage/migrations"
	"certwarden-backend/pkg/storage/sqlite3"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// config for DB
const dbTimeout = time.Duration(5 * time.Second)
const DbCurrentUserVersion = 11

var errServiceComponent = errors.New("necessary storage service component is missing")

type App interface {
	GetDataStorageAppDataPath() string
	GetLogger() *zap.SugaredLogger
	GetShutdownContext() context.Context
	CreateBackupOnDisk() error
}

// Storage is the db storage service
type Storage struct {
	shutdownContext context.Context
	db              *sql.DB
	timeout         time.Duration
}

// OpenStorage opens an existing sqlite database or creates a new one if needed.
// It also creates tables. It then returns Storage.
func OpenStorage(app App) (*Storage, error) {
	store := new(Storage)
	var err error

	// get shutdown context
	store.shutdownContext = app.GetShutdownContext()
	if store.shutdownContext == nil {
		return nil, errServiceComponent
	}

	// set timeout
	store.timeout = dbTimeout

	// logger just for setup
	logger := app.GetLogger()
	if logger == nil {
		return nil, errServiceComponent
	}

	db, cleanUpOnErr, err := sqlite3.OpenSqlite3Database(app)
	store.db = db
	if err != nil {
		return nil, fmt.Errorf("storage: failed to open database (%w)", err)
	}

	ctx, cancel := context.WithTimeout(store.shutdownContext, store.timeout)
	defer cancel()

	err = store.db.PingContext(ctx)
	if err != nil {
		cleanUpOnErr()
		return nil, fmt.Errorf("storage: failed to ping db after opening (%w)", err)
	}

	// perform any necessary db schema updates
	err = migrations.Do(ctx, db, logger)
	if err != nil {
		return nil, err
	}

	return store, nil
}

// Close() closes the storage database
func (store *Storage) Close() error {
	err := store.db.Close()
	if err != nil {
		return err
	}

	return nil
}
