package sqlite

import (
	"context"
	"database/sql"
	"log"
	"net/url"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// constant for DB timeout
const dbTimeout = time.Duration(5 * time.Second)

type Storage struct {
	Db      *sql.DB
	Timeout time.Duration
}

// function opens connection to the sqlite database
//   this will also cause the file to be created, if it does not exist
func OpenDB(dsn string, options url.Values) (*Storage, error) {
	storage := new(Storage)
	var err error

	// set timeout
	storage.Timeout = dbTimeout

	// append options to the Dsn
	connString := dsn + "?" + options.Encode()

	storage.Db, err = sql.Open("sqlite3", connString)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), storage.Timeout)
	defer cancel()

	err = storage.Db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	// create tables in the database if they don't exist
	err = storage.createDBTables()
	if err != nil {
		return nil, err
	}

	return storage, nil
}

// TODO - Close DB func for defer in main?

// function creates tables in the event our database is new
func (storage *Storage) createDBTables() error {
	ctx, cancel := context.WithTimeout(context.Background(), storage.Timeout)
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

	_, err := storage.Db.ExecContext(ctx, query)
	if err != nil {
		log.Fatal(err)
		return err
	}

	// acme_accounts
	query = `CREATE TABLE IF NOT EXISTS acme_accounts (
		id integer PRIMARY KEY AUTOINCREMENT NOT NULL UNIQUE,
		name text NOT NULL UNIQUE COLLATE NOCASE,
		private_key_id integer NOT NULL,
		description text,
		status text NOT NULL DEFAULT 'Unknown',
		email text NOT NULL,
		accepted_tos boolean DEFAULT 0,
		is_staging boolean DEFAULT 0,
		created_at datetime NOT NULL,
		updated_at datetime NOT NULL,
		kid text UNIQUE,
		FOREIGN KEY (private_key_id)
			REFERENCES private_keys (id)
				ON DELETE NO ACTION
				ON UPDATE NO ACTION
	)`

	_, err = storage.Db.ExecContext(ctx, query)
	if err != nil {
		log.Fatal(err)
		return err
	}

	// certificates
	query = `CREATE TABLE IF NOT EXISTS certificates (
		id integer PRIMARY KEY AUTOINCREMENT NOT NULL UNIQUE,
		private_key_id integer NOT NULL,
		acme_account_id integer NOT NULL,
		name text NOT NULL UNIQUE COLLATE NOCASE,
		description text,
		challenge_type integer NOT NULL,
		subject text NOT NULL,
		subject_alts text,
		csr_com_name text,
		csr_org text,
		csr_country text,
		csr_state text,
		csr_city text,
		created_at datetime NOT NULL,
		updated_at datetime NOT NULL,
		api_key text NOT NULL,
		pem text NOT NULL,
		valid_from datetime,
		valid_to datetime,
		is_valid boolean DEFAULT 0,
		FOREIGN KEY (private_key_id)
			REFERENCES private_keys (id)
				ON DELETE NO ACTION
				ON UPDATE NO ACTION,
		FOREIGN KEY (acme_account_id)
			REFERENCES acme_accounts (id)
				ON DELETE NO ACTION
				ON UPDATE NO ACTION
	)`

	_, err = storage.Db.ExecContext(ctx, query)
	if err != nil {
		log.Fatal(err)
		return err
	}

	return nil
}
