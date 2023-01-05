package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"legocerthub-backend/pkg/challenges"
	"legocerthub-backend/pkg/domain/app/auth"
	"net/url"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

var errServiceComponent = errors.New("necessary storage service component is missing")

// config for DB
const dbTimeout = time.Duration(5 * time.Second)
const dbDsn = "./lego-certhub.db"

var dbOptions = url.Values{
	"_fk": []string{"true"},
}

// App interface is for connecting to the main app
type App interface {
	GetChallengesService() *challenges.Service
}

// Storage is the struct that holds data about the connection
type Storage struct {
	Db         *sql.DB
	Timeout    time.Duration
	challenges *challenges.Service
}

// OpenStorage opens an existing sqlite database or creates a new one if needed.
// It also creates tables. It then returns Storage.
func OpenStorage(app App) (*Storage, error) {
	store := new(Storage)
	var err error

	// challenges service
	store.challenges = app.GetChallengesService()
	if store.challenges == nil {
		return nil, errServiceComponent
	}

	// set timeout
	store.Timeout = dbTimeout

	// append options to the Dsn
	connString := dbDsn + "?" + dbOptions.Encode()

	// check if db file exists
	dbExists := true
	if _, err := os.Stat(dbDsn); errors.Is(err, os.ErrNotExist) {
		dbExists = false
		// create db file
		file, err := os.Create(dbDsn)
		if err != nil {
			return nil, err
		}
		file.Close()
	}

	// open db
	store.Db, err = sql.Open("sqlite3", connString)
	if err != nil {
		// if db file is new, delete it on error
		if !dbExists {
			_ = store.Db.Close()
			_ = os.Remove(dbDsn)
		}
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), store.Timeout)
	defer cancel()

	err = store.Db.PingContext(ctx)
	if err != nil {
		// if db file is new, delete it on error
		if !dbExists {
			_ = store.Db.Close()
			_ = os.Remove(dbDsn)
		}
		return nil, err
	}

	// create tables in the database if the file is new
	if !dbExists {
		err = store.createDBTables()
		if err != nil {
			// delete new db on error setting it up
			_ = store.Db.Close()
			_ = os.Remove(dbDsn)
			return nil, err
		}
	}

	return store, nil
}

// Close() closes the storage database
func (store *Storage) Close() error {
	err := store.Db.Close()
	if err != nil {
		return err
	}

	return nil
}

// createDBTables creates tables in the event our database is new
func (store *Storage) createDBTables() error {
	ctx, cancel := context.WithTimeout(context.Background(), store.Timeout)
	defer cancel()

	// private_keys
	query := `CREATE TABLE private_keys (
		id integer PRIMARY KEY AUTOINCREMENT NOT NULL UNIQUE,
		name text NOT NULL UNIQUE COLLATE NOCASE,
		description text NOT NULL,
		algorithm text NOT NULL,
		pem text NOT NULL UNIQUE,
		api_key text NOT NULL,
		api_key_disabled NOT NULL DEFAULT 0 CHECK(api_key_via_url IN (0,1)),
		api_key_via_url integer NOT NULL DEFAULT 0 CHECK(api_key_via_url IN (0,1)),
		created_at integer NOT NULL,
		updated_at integer NOT NULL
	)`

	_, err := store.Db.ExecContext(ctx, query)
	if err != nil {
		return err
	}

	// acme_accounts
	query = `CREATE TABLE acme_accounts (
		id integer PRIMARY KEY AUTOINCREMENT NOT NULL UNIQUE,
		name text NOT NULL UNIQUE COLLATE NOCASE,
		private_key_id integer NOT NULL UNIQUE,
		description text NOT NULL,
		status text NOT NULL DEFAULT 'unknown',
		email text NOT NULL,
		accepted_tos boolean NOT NULL DEFAULT 0,
		is_staging boolean NOT NULL DEFAULT 0,
		created_at integer NOT NULL,
		updated_at integer NOT NULL,
		kid text NOT NULL,
		FOREIGN KEY (private_key_id)
			REFERENCES private_keys (id)
				ON DELETE RESTRICT
				ON UPDATE NO ACTION
	)`

	_, err = store.Db.ExecContext(ctx, query)
	if err != nil {
		return err
	}

	// certificates
	query = `CREATE TABLE certificates (
		id integer PRIMARY KEY AUTOINCREMENT NOT NULL UNIQUE,
		private_key_id integer NOT NULL UNIQUE,
		acme_account_id integer NOT NULL,
		name text NOT NULL UNIQUE COLLATE NOCASE,
		description text NOT NULL,
		challenge_method text NOT NULL,
		subject text NOT NULL,
		subject_alts text NOT NULL,
		csr_org text NOT NULL,
		csr_ou text NOT NULL,
		csr_country text NOT NULL,
		csr_state text NOT NULL,
		csr_city text NOT NULL,
		api_key text NOT NULL,
		api_key_via_url integer NOT NULL DEFAULT 0 CHECK(api_key_via_url IN (0,1)),
		created_at integer NOT NULL,
		updated_at integer NOT NULL,
		FOREIGN KEY (private_key_id)
			REFERENCES private_keys (id)
				ON DELETE RESTRICT
				ON UPDATE NO ACTION,
		FOREIGN KEY (acme_account_id)
			REFERENCES acme_accounts (id)
				ON DELETE RESTRICT
				ON UPDATE NO ACTION
	)`

	_, err = store.Db.ExecContext(ctx, query)
	if err != nil {
		return err
	}

	// ACME orders
	query = `CREATE TABLE acme_orders (
			id integer PRIMARY KEY AUTOINCREMENT NOT NULL UNIQUE,
			acme_account_id integer NOT NULL,
			certificate_id integer NOT NULL,
			acme_location text NOT NULL UNIQUE,
			status text NOT NULL,
			known_revoked integer NOT NULL DEFAULT 0 CHECK(known_revoked IN (0,1)),
			error text,
			expires integer,
			dns_identifiers text NOT NULL,
			authorizations text NOT NULL,
			finalize text NOT NULL,
			finalized_key_id integer,
			certificate_url text,
			pem text,
			valid_from integer,
			valid_to integer,
			created_at integer NOT NULL,
			updated_at integer NOT NULL,
			FOREIGN KEY (acme_account_id)
				REFERENCES acme_accounts (id)
					ON DELETE CASCADE
					ON UPDATE NO ACTION,
			FOREIGN KEY (finalized_key_id)
				REFERENCES private_keys (id)
					ON DELETE SET NULL
					ON UPDATE NO ACTION,
			FOREIGN KEY (certificate_id)
				REFERENCES certificates (id)
					ON DELETE CASCADE
					ON UPDATE NO ACTION
		)`

	_, err = store.Db.ExecContext(ctx, query)
	if err != nil {
		return err
	}

	// users (for login to LeGo)
	query = `CREATE TABLE users (
		id integer PRIMARY KEY AUTOINCREMENT NOT NULL UNIQUE,
		username text NOT NULL UNIQUE,
		password_hash NOT NULL,
		created_at integer NOT NULL,
		updated_at integer NOT NULL
	)`

	_, err = store.Db.ExecContext(ctx, query)
	if err != nil {
		return err
	}

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
	query = `
	INSERT INTO
		users (username, password_hash, created_at, updated_at)
	VALUES (
		$1,
		$2,
		$3,
		$4
	)
	`

	_, err = store.Db.ExecContext(ctx, query,
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
