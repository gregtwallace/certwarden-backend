package sqlite

import (
	"context"
	"database/sql"
	"log"
	"net/url"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// config for DB
const dbTimeout = time.Duration(5 * time.Second)
const dbDsn = "./lego-certhub.db"

var dbOptions = url.Values{
	"_fk": []string{"true"},
}

// Storage is the struct that holds data about the connection
type Storage struct {
	Db      *sql.DB
	Timeout time.Duration
}

// OpenStorage opens an existing sqlite database or creates a new one if needed.
//   It also creates tables. It then returns Storage.
func OpenStorage() (*Storage, error) {
	store := new(Storage)
	var err error

	// set timeout
	store.Timeout = dbTimeout

	// append options to the Dsn
	connString := dbDsn + "?" + dbOptions.Encode()

	store.Db, err = sql.Open("sqlite3", connString)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), store.Timeout)
	defer cancel()

	err = store.Db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	// create tables in the database if they don't exist
	err = store.createDBTables()
	if err != nil {
		return nil, err
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
	query := `CREATE TABLE IF NOT EXISTS private_keys (
		id integer PRIMARY KEY AUTOINCREMENT NOT NULL UNIQUE,
		name text NOT NULL UNIQUE COLLATE NOCASE,
		description text,
		algorithm text NOT NULL,
		pem text NOT NULL UNIQUE,
		api_key text NOT NULL,
		created_at integer NOT NULL,
		updated_at integer NOT NULL
	)`

	_, err := store.Db.ExecContext(ctx, query)
	if err != nil {
		log.Fatal(err)
		return err
	}

	// acme_accounts
	query = `CREATE TABLE IF NOT EXISTS acme_accounts (
		id integer PRIMARY KEY AUTOINCREMENT NOT NULL UNIQUE,
		name text NOT NULL UNIQUE COLLATE NOCASE,
		private_key_id integer NOT NULL UNIQUE,
		description text,
		status text NOT NULL DEFAULT 'Unknown',
		email text NOT NULL,
		accepted_tos boolean DEFAULT 0,
		is_staging boolean DEFAULT 0,
		created_at integer NOT NULL,
		updated_at integer NOT NULL,
		kid text UNIQUE,
		FOREIGN KEY (private_key_id)
			REFERENCES private_keys (id)
				ON DELETE RESTRICT
				ON UPDATE NO ACTION
	)`

	_, err = store.Db.ExecContext(ctx, query)
	if err != nil {
		log.Fatal(err)
		return err
	}

	// certificates
	query = `CREATE TABLE IF NOT EXISTS certificates (
		id integer PRIMARY KEY AUTOINCREMENT NOT NULL UNIQUE,
		private_key_id integer NOT NULL UNIQUE,
		acme_account_id integer NOT NULL,
		name text NOT NULL UNIQUE COLLATE NOCASE,
		description text,
		challenge_method text NOT NULL,
		subject text NOT NULL,
		subject_alts text,
		csr_org text,
		csr_ou text,
		csr_country text,
		csr_state text,
		csr_city text,
		api_key text NOT NULL,
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
		log.Fatal(err)
		return err
	}

	// ACME orders
	query = `CREATE TABLE IF NOT EXISTS acme_orders (
			id integer PRIMARY KEY AUTOINCREMENT NOT NULL UNIQUE,
			acme_account_id integer NOT NULL,
			certificate_id integer,
			acme_location text NOT NULL UNIQUE,
			status text NOT NULL,
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
					ON DELETE SET NULL
					ON UPDATE NO ACTION
		)`

	_, err = store.Db.ExecContext(ctx, query)
	if err != nil {
		log.Fatal(err)
		return err
	}

	return nil
}
