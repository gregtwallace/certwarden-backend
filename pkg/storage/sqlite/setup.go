package sqlite

import (
	"certwarden-backend/pkg/domain/app/auth"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// config for DB
const dbTimeout = time.Duration(5 * time.Second)
const DbFilename = "appdata.db"
const DbCurrentUserVersion = 7
const dbFileMode = 0600

var dbOptions = url.Values{
	"_fk": []string{"true"},
}

var errServiceComponent = errors.New("necessary storage service component is missing")

type App interface {
	GetDataStorageAppDataPath() string
	GetLogger() *zap.SugaredLogger
	CreateBackupOnDisk() error
}

// Storage is the struct that holds data about the connection
type Storage struct {
	logger  *zap.SugaredLogger
	db      *sql.DB
	timeout time.Duration
}

// OpenStorage opens an existing sqlite database or creates a new one if needed.
// It also creates tables. It then returns Storage.
func OpenStorage(app App) (*Storage, error) {
	store := new(Storage)
	var err error

	// get logger
	store.logger = app.GetLogger()
	if store.logger == nil {
		return nil, errServiceComponent
	}

	// set timeout
	store.timeout = dbTimeout

	// full path and append options to the Dsn for connString
	dbWithPath := app.GetDataStorageAppDataPath() + "/" + DbFilename
	connString := dbWithPath + "?" + dbOptions.Encode()

	// check if db file exists
	dbExists := true
	if _, err := os.Stat(dbWithPath); errors.Is(err, os.ErrNotExist) {
		// if doesn't exist, check for old db file name (from pre-rename)
		oldDbWithPath := app.GetDataStorageAppDataPath() + "/lego-certhub.db"
		_, err := os.Stat(oldDbWithPath)
		if err != nil {
			// neither new or old file name exist
			dbExists = false
			store.logger.Warn("database file does not exist, creating a new one")
			// create db file
			err := os.WriteFile(dbWithPath, []byte{}, dbFileMode)
			if err != nil {
				store.logger.Errorf("failed to create new database file", err)
				return nil, err
			}
		} else {
			// rename old db to new filename
			err = os.Rename(oldDbWithPath, dbWithPath)
			if err != nil {
				store.logger.Errorf("failed to rename old database file to new filename", err)
				return nil, err
			}
			store.logger.Info("old database file renamed to new database filename")
		}
	}

	// open db
	store.db, err = sql.Open("sqlite3", connString)
	if err != nil {
		// if db file is new, delete it on error
		if !dbExists {
			_ = store.db.Close()
			_ = os.Remove(dbWithPath)
		}
		store.logger.Errorf("failed to open database file", err)
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), store.timeout)
	defer cancel()

	err = store.db.PingContext(ctx)
	if err != nil {
		// if db file is new, delete it on error
		if !dbExists {
			_ = store.db.Close()
			_ = os.Remove(dbWithPath)
		}
		store.logger.Errorf("failed to ping database file after opening", err)
		return nil, err
	}

	// create tables in the database if the file is new
	if !dbExists {
		store.logger.Info("populating new database file")
		err = store.populateNewDb()
		if err != nil {
			// delete new db on error setting it up
			_ = store.db.Close()
			_ = os.Remove(dbWithPath)
			store.logger.Errorf("failed to populate new database file", err)
			return nil, err
		}
	}

	// check and do db schema upgrades, if needed
	// get db file user_version
	query := `PRAGMA user_version`
	row := store.db.QueryRowContext(ctx, query)
	fileUserVersion := -1
	err = row.Scan(
		&fileUserVersion,
	)
	if err != nil {
		return nil, err
	}

	// back up exisitng db before trying any migrations
	if fileUserVersion != DbCurrentUserVersion {
		err = app.CreateBackupOnDisk()
		if err != nil {
			return nil, fmt.Errorf("failed to backup data before attempting db migration (%s)", err)
		}
	}

	// upgrade if schema 0
	if fileUserVersion == 0 {
		fileUserVersion, err = store.migrateV0toV1()
		if err != nil {
			return nil, err
		}
	}

	// upgrade if schema 1
	if fileUserVersion == 1 {
		fileUserVersion, err = store.migrateV1toV2()
		if err != nil {
			return nil, err
		}
	}

	// upgrade if schema 2
	if fileUserVersion == 2 {
		fileUserVersion, err = store.migrateV2toV3()
		if err != nil {
			return nil, err
		}
	}

	// upgrade if schema 3
	if fileUserVersion == 3 {
		fileUserVersion, err = store.migrateV3toV4()
		if err != nil {
			return nil, err
		}
	}

	// upgrade if schema 4
	if fileUserVersion == 4 {
		fileUserVersion, err = store.migrateV4toV5()
		if err != nil {
			return nil, err
		}
	}

	// upgrade if schema 5
	if fileUserVersion == 5 {
		fileUserVersion, err = store.migrateV5toV6()
		if err != nil {
			return nil, err
		}
	}

	// upgrade if schema 6
	if fileUserVersion == 6 {
		fileUserVersion, err = store.migrateV6toV7()
		if err != nil {
			return nil, err
		}
	}

	// fail if still not correct
	if fileUserVersion != DbCurrentUserVersion {
		return nil, fmt.Errorf("db schema user_version is %d (expected %d) and automatic migration failed", fileUserVersion, DbCurrentUserVersion)
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

// populateNewDb creates the tables in the db file and sets the db version
func (store *Storage) populateNewDb() error {
	ctx, cancel := context.WithTimeout(context.Background(), store.timeout)
	defer cancel()

	// create sql transaction to roll back in the event an error occurs
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// set db user_version
	// No injection protection since const isn't user editable
	query := `PRAGMA user_version = ` + strconv.Itoa(DbCurrentUserVersion)

	_, err = tx.ExecContext(ctx, query)
	if err != nil {
		return err
	}

	// create tables
	err = createDBTablesV7(tx)
	if err != nil {
		return err
	}

	// insert default ACME servers
	err = insertDefaultAcmeServers(tx)
	if err != nil {
		return err
	}

	// insert default user
	err = insertDefaultUser(tx)
	if err != nil {
		return err
	}

	// no errors, commit transaction
	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

// insertDefaultUser inserts the default user with the default password
func insertDefaultUser(tx *sql.Tx) error {
	// add default admin user
	// default username and password
	defaultUsername := "admin"
	defaultPassword := "password"

	// generate password hash
	defaultHashedPw, err := bcrypt.GenerateFromPassword([]byte(defaultPassword), auth.BcryptCost)
	if err != nil {
		return err
	}

	// insert
	query := `
	INSERT INTO
		users (username, password_hash, created_at, updated_at)
	VALUES (
		$1,
		$2,
		$3,
		$4
	)
	`

	_, err = tx.Exec(query,
		defaultUsername,
		defaultHashedPw,
		time.Now().Unix(),
		time.Now().Unix(),
	)

	if err != nil {
		return err
	}

	return nil
}
