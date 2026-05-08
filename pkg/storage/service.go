package storage

import (
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

	// set timeout
	store.timeout = dbTimeout

	db, isNewDb, cleanUpOnErr, err := sqlite3.OpenSqlite3Database(app)
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

	if isNewDb {
		app.GetLogger().Info("storage: populating new database")
		err = store.populateNewDb()
		if err != nil {
			cleanUpOnErr()
			return nil, fmt.Errorf("storage: failed to populate new database (%w)", err)
		}
	}

	// check and do db schema upgrades, if needed
	// get db file user_version
	query := `PRAGMA user_version`
	row := store.db.QueryRowContext(ctx, query)
	fileOriginalUserVersion := -1
	err = row.Scan(
		&fileOriginalUserVersion,
	)
	if err != nil {
		cleanUpOnErr()
		return nil, err
	}

	// back up exisitng db before trying any migrations
	if fileOriginalUserVersion != DbCurrentUserVersion {
		err = app.CreateBackupOnDisk()
		if err != nil {
			cleanUpOnErr()
			return nil, fmt.Errorf("storage: failed to backup data before attempting db migration (%w)", err)
		}
		app.GetLogger().Infof("storage: updating database user_version from %d to %d", fileOriginalUserVersion, DbCurrentUserVersion)
	}

	// upgrade if schema 0
	fileUserVersion := fileOriginalUserVersion
	if fileUserVersion == 0 {
		fileUserVersion, err = store.migrateV0toV1()
		if err != nil {
			cleanUpOnErr()
			return nil, fmt.Errorf("storage: failed to migrate to user_version %d (%w)", fileUserVersion+1, err)
		}
	}

	// upgrade if schema 1
	if fileUserVersion == 1 {
		fileUserVersion, err = store.migrateV1toV2()
		if err != nil {
			cleanUpOnErr()
			return nil, fmt.Errorf("storage: failed to migrate to user_version %d (%w)", fileUserVersion+1, err)
		}
	}

	// upgrade if schema 2
	if fileUserVersion == 2 {
		fileUserVersion, err = store.migrateV2toV3()
		if err != nil {
			cleanUpOnErr()
			return nil, fmt.Errorf("storage: failed to migrate to user_version %d (%w)", fileUserVersion+1, err)
		}
	}

	// upgrade if schema 3
	if fileUserVersion == 3 {
		fileUserVersion, err = store.migrateV3toV4()
		if err != nil {
			cleanUpOnErr()
			return nil, fmt.Errorf("storage: failed to migrate to user_version %d (%w)", fileUserVersion+1, err)
		}
	}

	// upgrade if schema 4
	if fileUserVersion == 4 {
		fileUserVersion, err = store.migrateV4toV5()
		if err != nil {
			cleanUpOnErr()
			return nil, fmt.Errorf("storage: failed to migrate to user_version %d (%w)", fileUserVersion+1, err)
		}
	}

	// upgrade if schema 5
	if fileUserVersion == 5 {
		fileUserVersion, err = store.migrateV5toV6()
		if err != nil {
			cleanUpOnErr()
			return nil, fmt.Errorf("storage: failed to migrate to user_version %d (%w)", fileUserVersion+1, err)
		}
	}

	// upgrade if schema 6
	if fileUserVersion == 6 {
		fileUserVersion, err = store.migrateV6toV7()
		if err != nil {
			cleanUpOnErr()
			return nil, fmt.Errorf("storage: failed to migrate to user_version %d (%w)", fileUserVersion+1, err)
		}
	}

	// upgrade if schema 7
	if fileUserVersion == 7 {
		fileUserVersion, err = store.migrateV7toV8()
		if err != nil {
			cleanUpOnErr()
			return nil, fmt.Errorf("storage: failed to migrate to user_version %d (%w)", fileUserVersion+1, err)
		}
	}

	// upgrade if schema 8
	if fileUserVersion == 8 {
		fileUserVersion, err = store.migrateV8toV9()
		if err != nil {
			cleanUpOnErr()
			return nil, fmt.Errorf("storage: failed to migrate to user_version %d (%w)", fileUserVersion+1, err)
		}
	}

	// upgrade if schema 9
	if fileUserVersion == 9 {
		fileUserVersion, err = store.migrateV9toV10()
		if err != nil {
			cleanUpOnErr()
			return nil, fmt.Errorf("storage: failed to migrate to user_version %d (%w)", fileUserVersion+1, err)
		}
	}

	// upgrade if schema 10
	if fileUserVersion == 10 {
		fileUserVersion, err = store.migrateV10toV11()
		if err != nil {
			cleanUpOnErr()
			return nil, fmt.Errorf("storage: failed to migrate to user_version %d (%w)", fileUserVersion+1, err)
		}
	}

	// fail if still not correct
	if fileUserVersion != DbCurrentUserVersion {
		cleanUpOnErr()
		return nil, fmt.Errorf("storage: db schema user_version is %d (expected %d) and automatic migration failed", fileUserVersion, DbCurrentUserVersion)
	}
	if fileOriginalUserVersion != DbCurrentUserVersion {
		app.GetLogger().Infof("storage: database user_version successfully upgraded from %d to %d", fileOriginalUserVersion, fileUserVersion)
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
