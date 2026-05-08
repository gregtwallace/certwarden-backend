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
const dbFileMode = 0600

var dbOptions = url.Values{
	"_fk": []string{"true"},
}

type App interface {
	GetLogger() *zap.SugaredLogger
	GetDataStorageAppDataPath() string
}

// OpenSqlite3Database
func OpenSqlite3Database(app App) (_ *sql.DB, isNewDb bool, onErrCleanup func(), _ error) {
	// full path and append options to the Dsn for connString
	dbWithPath := app.GetDataStorageAppDataPath() + "/" + dbFilename
	connString := dbWithPath + "?" + dbOptions.Encode()

	// check if db file exists
	newDbFile := false
	if _, err := os.Stat(dbWithPath); errors.Is(err, os.ErrNotExist) {
		// db doesn't exist, check old path
		didMigrate, err := migrateDbFileLocation(dbWithPath)
		if err != nil {
			return nil, false, func() {}, fmt.Errorf("sqlite3: db migration failed (%w)", err)
		}
		if didMigrate {
			// old db migrated
			app.GetLogger().Infof("sqlite3: db file moved to %s", dbWithPath)
		} else {
			// new db
			newDbFile = true
			app.GetLogger().Warn("sqlite3: database file does not exist, creating a new one")
			// create db file
			err := os.WriteFile(dbWithPath, []byte{}, dbFileMode)
			if err != nil {
				return nil, false, func() {}, fmt.Errorf("sqlite3: failed to create new database file (%w)", err)
			}
		}
	}

	// open db
	db, err := sql.Open("sqlite3", connString)
	if err != nil {
		// if db file is new, delete it on error
		if newDbFile {
			_ = os.Remove(dbWithPath)
		}
		return nil, false, func() {}, fmt.Errorf("sqlite3: failed to open database file (%w)", err)
	}

	// cleanup func for if there is a failure later
	cleanUp := func() {
		_ = db.Close()

		// only remove file if new file was created
		if newDbFile {
			_ = os.Remove(dbWithPath)
		}
	}

	return db, newDbFile, cleanUp, nil
}

// pre-migration file path
const oldFile = "./data/" + dbFilename

// migrateDbFileLocation moves the db file from its "old" location to the current one
func migrateDbFileLocation(migrateToFile string) (didMigrate bool, _ error) {
	// stat old location
	_, err := os.Stat(oldFile)
	if err != nil {
		// no old file
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		// any other error
		return false, fmt.Errorf("sqlite3: could not check for db file at old location (%w)", err)
	}

	// old file exists
	err = os.Rename(oldFile, migrateToFile)
	if err != nil {
		return false, fmt.Errorf("sqlite3: failed to move existing db file from %s to %s (%s)", oldFile, migrateToFile, err)
	}

	return true, nil
}
