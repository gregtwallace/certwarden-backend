package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"legocerthub-backend/pkg/domain/app/auth"
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
const dbFilename = "/lego-certhub.db"
const DbCurrentUserVersion = 1

var dbOptions = url.Values{
	"_fk": []string{"true"},
}

var errServiceComponent = errors.New("necessary storage service component is missing")

type App interface {
	GetLogger() *zap.SugaredLogger
}

// Storage is the struct that holds data about the connection
type Storage struct {
	logger  *zap.SugaredLogger
	db      *sql.DB
	timeout time.Duration
}

// OpenStorage opens an existing sqlite database or creates a new one if needed.
// It also creates tables. It then returns Storage.
func OpenStorage(app App, dataPath string) (*Storage, error) {
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
	dbWithPath := dataPath + dbFilename
	connString := dbWithPath + "?" + dbOptions.Encode()

	// check if db file exists
	dbExists := true
	if _, err := os.Stat(dbWithPath); errors.Is(err, os.ErrNotExist) {
		dbExists = false
		store.logger.Warn("database file does not exist, creating a new one")
		// create db file
		file, err := os.Create(dbWithPath)
		if err != nil {
			store.logger.Errorf("failed to create new database file", err)
			return nil, err
		}
		file.Close()
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
	} else {
		// check and do db schema upgrades, if needed
		fileUserVersion := -1
		// try upgrading until version matches or an error occurs
		for fileUserVersion != DbCurrentUserVersion && err == nil {
			// get db file user_version
			query := `PRAGMA user_version`
			row := store.db.QueryRowContext(ctx, query)
			err = row.Scan(
				&fileUserVersion,
			)
			if err != nil {
				return nil, err
			}

			// take incremental migration action, if needed
			switch fileUserVersion {
			case 0:
				err = store.migrateV0toV1()
			case DbCurrentUserVersion:
				store.logger.Debugf("database user_version is current (%d)", fileUserVersion)
				// no-op, loop will end due to version ==
			default:
				err = errors.New("unsupported user_version found in db file")
			}
		}

		// err check from upgrade loop
		if err != nil {
			store.logger.Errorf("failed to update database file to latest user_version (%d), currently %d (%s)", DbCurrentUserVersion, fileUserVersion, err)
			return nil, err
		}

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
	err = createDBTables(tx)
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
