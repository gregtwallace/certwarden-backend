package main

import (
	"context"
	"database/sql"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type DBWrap struct {
	DB *sql.DB
}

// function opens connection to the sqlite database
//   this will also cause the file to be created, if it does not exist
func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}

// function creates tables in the event our database is new
func (db *DBWrap) createDBTables() error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// private_keys
	query := `CREATE TABLE IF NOT EXISTS private_keys (
		id integer PRIMARY KEY AUTOINCREMENT NOT NULL UNIQUE,
		name text NOT NULL UNIQUE,
		description text,
		algorithm text NOT NULL,
		pem_content text NOT NULL,
		api_key text NOT NULL,
		created_at text NOT NULL,
		updated_at text NOT NULL
	)`

	_, err := db.DB.ExecContext(ctx, query)
	if err != nil {
		log.Fatal(err)
		return err
	}

	// acme_accounts
	query = `CREATE TABLE IF NOT EXISTS acme_accounts (
		id integer PRIMARY KEY AUTOINCREMENT NOT NULL UNIQUE,
		private_key_id integer NOT NULL,
		name text NOT NULL UNIQUE,
		description text,
		email text NOT NULL,
		accepted_tos integer DEFAULT 0,
		is_staging integer DEFAULT 0,
		created_at text NOT NULL,
		updated_at text NOT NULL,
		FOREIGN KEY (private_key_id)
			REFERENCES private_keys (id)
				ON DELETE CASCADE
				ON UPDATE NO ACTION
	)`

	_, err = db.DB.ExecContext(ctx, query)
	if err != nil {
		log.Fatal(err)
		return err
	}

	// certificates
	query = `CREATE TABLE IF NOT EXISTS certificates (
		id integer PRIMARY KEY AUTOINCREMENT NOT NULL UNIQUE,
		private_key_id integer NOT NULL,
		acme_account_id integer NOT NULL,
		name text NOT NULL UNIQUE,
		description text,
		challenge_type integer NOT NULL,
		subject text NOT NULL,
		subject_alts text,
		csr_com_name text,
		csr_org text,
		csr_country text,
		csr_state text,
		csr_city text,
		created_at text NOT NULL,
		updated_at text NOT NULL,
		api_key text NOT NULL,
		valid_from text,
		valid_to text,
		is_valid integer DEFAULT 0,
		FOREIGN KEY (private_key_id)
			REFERENCES private_keys (id)
				ON DELETE CASCADE
				ON UPDATE NO ACTION,
		FOREIGN KEY (acme_account_id)
			REFERENCES acme_accounts (id)
				ON DELETE CASCADE
				ON UPDATE NO ACTION
	)`

	_, err = db.DB.ExecContext(ctx, query)
	if err != nil {
		log.Fatal(err)
		return err
	}

	return nil
}
