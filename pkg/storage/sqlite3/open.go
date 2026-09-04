package sqlite3

import (
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"

	"go.uber.org/zap"

	_ "github.com/mattn/go-sqlite3"
)

const dbFilename = "appdata.db"
const dbFileMode = 0o600

var dbOptions = url.Values{
	"_fk": []string{"true"}, // enforce foreign key constraints
}

type App interface {
	GetLogger() *zap.SugaredLogger
	GetDataStorageAppDataPath() string
}

// OpenSqlite3Database
func OpenSqlite3Database(app App) (_ *sql.DB, onErrCleanup func(), _ error) {
	// full path and append options to the Dsn for connString
	dbWithPath := app.GetDataStorageAppDataPath() + "/" + dbFilename

	logger := app.GetLogger()

	// check if db file exists
	dbExists := true
	newDbFile := false
	_, err := os.Stat(dbWithPath)
	if errors.Is(err, os.ErrNotExist) {
		dbExists = false
	} else if err != nil {
		// any other error
		return nil, func() {}, fmt.Errorf("sqlite3: failed to stat db (%w) (%w)", err, errStatToFailed)
	}

	// db doesn't exist, check old path
	if !dbExists {
		didMigrate, err := migrateDbFileLocation(oldFilePath+"/"+dbFilename, dbWithPath)
		if err != nil {
			return nil, func() {}, fmt.Errorf("sqlite3: db migration failed (%w)", err)
		}
		if didMigrate {
			// old db migrated
			logger.Infof("sqlite3: db file moved to %s", dbWithPath)
		} else {
			// no old one either, so new db
			newDbFile = true
		}
	}

	// setup new db
	if newDbFile {
		logger.Warn("sqlite3: database file does not exist, creating a new one")
		// create db file
		err := os.WriteFile(dbWithPath, []byte{}, dbFileMode)
		if err != nil {
			return nil, func() {}, fmt.Errorf("sqlite3: failed to create new database file (%w)", err)
		}
	}

	// open db
	db, err := sql.Open("sqlite3", dbWithPath+"?"+dbOptions.Encode())
	if err != nil {
		// if db file is new, delete it on error
		if newDbFile {
			err := os.Remove(dbWithPath)
			if err != nil {
				logger.Errorf("sqlite3: failed to remove newly created db file (%s)", err)
			}
		}
		return nil, func() {}, fmt.Errorf("sqlite3: failed to open database file (%w)", err)
	}

	// cleanup func for if there is a failure later
	onErrCleanup = func() {
		//nolint:errcheck // who cares, app already failed anyway
		db.Close()

		// only remove file if new file was created
		if newDbFile {
			err := os.Remove(dbWithPath)
			if err != nil {
				logger.Errorf("sqlite3: failed to remove newly created db file (%s)", err)
			}
		}
	}

	return db, onErrCleanup, nil
}
