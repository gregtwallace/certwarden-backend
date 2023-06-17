package sqlite

import (
	"database/sql"
	"time"
)

// createDBTables creates a fresh set of tables in the db. This is used for new db
// file creation and for recreation during migration. This is also used to ensure
// all expected tables exist before renaming them to _old
func createDBTables(tx *sql.Tx) error {
	// acme_servers
	query := `CREATE TABLE IF NOT EXISTS acme_servers (
		id integer PRIMARY KEY AUTOINCREMENT NOT NULL UNIQUE,
		name text NOT NULL UNIQUE COLLATE NOCASE,
		description text NOT NULL,
		directory_url text NOT NULL UNIQUE,
		is_staging integer NOT NULL DEFAULT 0 CHECK(is_staging IN (0,1)),
		created_at integer NOT NULL,
		updated_at integer NOT NULL
	)`

	_, err := tx.Exec(query)
	if err != nil {
		return err
	}

	// private_keys
	query = `CREATE TABLE IF NOT EXISTS private_keys (
		id integer PRIMARY KEY AUTOINCREMENT NOT NULL UNIQUE,
		name text NOT NULL UNIQUE COLLATE NOCASE,
		description text NOT NULL,
		algorithm text NOT NULL,
		pem text NOT NULL UNIQUE,
		api_key text NOT NULL,
		api_key_new text NOT NULL DEFAULT '',
		api_key_disabled integer NOT NULL DEFAULT 0 CHECK(api_key_disabled IN (0,1)),
		api_key_via_url integer NOT NULL DEFAULT 0 CHECK(api_key_via_url IN (0,1)),
		created_at integer NOT NULL,
		updated_at integer NOT NULL
	)`

	_, err = tx.Exec(query)
	if err != nil {
		return err
	}

	// acme_accounts
	query = `CREATE TABLE IF NOT EXISTS acme_accounts (
		id integer PRIMARY KEY AUTOINCREMENT NOT NULL UNIQUE,
		name text NOT NULL UNIQUE COLLATE NOCASE,
		private_key_id integer NOT NULL UNIQUE,
		description text NOT NULL,
		status text NOT NULL DEFAULT 'unknown',
		email text NOT NULL,
		accepted_tos integer NOT NULL DEFAULT 0 CHECK(accepted_tos IN (0,1)),
		created_at integer NOT NULL,
		updated_at integer NOT NULL,
		kid text NOT NULL,
		acme_server_id integer NOT NULL,
		FOREIGN KEY (private_key_id)
			REFERENCES private_keys (id)
				ON DELETE RESTRICT
				ON UPDATE NO ACTION,
		FOREIGN KEY (acme_server_id)
			REFERENCES acme_servers (id)
				ON DELETE RESTRICT
				ON UPDATE NO ACTION
	)`

	_, err = tx.Exec(query)
	if err != nil {
		return err
	}

	// certificates
	query = `CREATE TABLE IF NOT EXISTS certificates (
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
		api_key_new text NOT NULL DEFAULT '',
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

	_, err = tx.Exec(query)
	if err != nil {
		return err
	}

	// ACME orders
	query = `CREATE TABLE IF NOT EXISTS acme_orders (
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

	_, err = tx.Exec(query)
	if err != nil {
		return err
	}

	// users (for login to LeGo)
	query = `CREATE TABLE IF NOT EXISTS users (
		id integer PRIMARY KEY AUTOINCREMENT NOT NULL UNIQUE,
		username text NOT NULL UNIQUE,
		password_hash NOT NULL,
		created_at integer NOT NULL,
		updated_at integer NOT NULL
	)`

	_, err = tx.Exec(query)
	if err != nil {
		return err
	}

	return nil
}

// insertDefaultAcmeServers inserts the default ACME Servers which are Let's Encrypt
// and Let's Encrypt (Staging)
func insertDefaultAcmeServers(tx *sql.Tx) error {
	// add LE acme servers
	query := `
		INSERT INTO acme_servers (id, name, description, directory_url, is_staging, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	// LE Prod
	_, err := tx.Exec(query,
		0,
		"Lets_Encrypt",
		"Let's Encrypt Production Server",
		"https://acme-v02.api.letsencrypt.org/directory",
		false,
		int(time.Now().Unix()),
		int(time.Now().Unix()),
	)
	if err != nil {
		return err
	}

	// LE Staging
	_, err = tx.Exec(query,
		1,
		"Lets_Encrypt_Staging",
		"Let's Encrypt Staging Server",
		"https://acme-staging-v02.api.letsencrypt.org/directory",
		true,
		int(time.Now().Unix()),
		int(time.Now().Unix()),
	)
	if err != nil {
		return err
	}

	return nil
}

// renameOldDbTables renames the existing tables to append _old to the end
// in preparation for re-creation of new tables. If any expected tables don't
// exist, they are created blank and then renamed to _old (this happens when
// a db version adds a new table)
func renameOldDbTables(tx *sql.Tx) error {
	// ensure all tables actually exist
	err := createDBTables(tx)
	if err != nil {
		return err
	}

	// rename tables
	query := `
		ALTER TABLE acme_accounts RENAME TO acme_accounts_old;
		ALTER TABLE acme_orders RENAME TO acme_orders_old;
		ALTER TABLE acme_servers RENAME TO acme_servers_old;
		ALTER TABLE certificates RENAME TO certificates_old;
		ALTER TABLE private_keys RENAME TO private_keys_old;
		ALTER TABLE users RENAME TO users_old;
	`

	_, err = tx.Exec(query)
	if err != nil {
		return err
	}

	return nil
}

// removeOldDbTables drops all of the _old table variants as a cleanup step
func removeOldDbTables(tx *sql.Tx) error {
	// drop tables
	query := `
		DROP TABLE acme_orders_old;	
		DROP TABLE certificates_old;
		DROP TABLE acme_accounts_old;
		DROP TABLE private_keys_old;
		DROP TABLE acme_servers_old;
		DROP TABLE users_old;
	`

	_, err := tx.Exec(query)
	if err != nil {
		return err
	}

	return nil
}
