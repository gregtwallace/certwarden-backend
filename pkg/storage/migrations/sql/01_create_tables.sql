-- NOTE: Support for automatic migration from v0 to v1 has been dropped!

-- ALSO NOTE: goose drops the use of pragma (as it is no longer needed and sqlite specific)

-- CHANGES v0 to v1:
-- - `acme_servers`
--     - Add table and attributes
--     - Add 2x entries to table for LE Prod (0) and LE Staging (1)
-- - `acme_accounts`:
--     - Add `acme_server_id` attribute
--     - Copy `is_staging` to `acme_server_id` (matches to 2x entries above without need
--       to change value)
--     - Remove `is_staging` attribute


-- +goose Up


-- new data tables
CREATE TABLE acme_servers (
	id integer PRIMARY KEY AUTOINCREMENT NOT NULL UNIQUE,
	name text NOT NULL UNIQUE COLLATE NOCASE,
	description text NOT NULL,
	directory_url text NOT NULL UNIQUE,
	is_staging integer NOT NULL DEFAULT 0 CHECK(is_staging IN (0,1)),
	created_at integer NOT NULL,
	updated_at integer NOT NULL
);

CREATE TABLE private_keys (
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
);

CREATE TABLE acme_accounts (
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
);

CREATE TABLE certificates (
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
);

CREATE TABLE acme_orders (
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
);

CREATE TABLE users (
	id integer PRIMARY KEY AUTOINCREMENT NOT NULL UNIQUE,
	username text NOT NULL UNIQUE,
	password_hash NOT NULL,
	created_at integer NOT NULL,
	updated_at integer NOT NULL
);

-- insert default servers
INSERT INTO acme_servers 
		(id, name, description, directory_url, is_staging, created_at, updated_at)
	VALUES (
		0,
		"Lets_Encrypt",
		"Let's Encrypt Production Server",
		"https://acme-v02.api.letsencrypt.org/directory",
		0,
		strftime('%s', 'now'),
		strftime('%s', 'now')
);

INSERT INTO acme_servers
		(id, name, description, directory_url, is_staging, created_at, updated_at)
	VALUES (
		1,
		"Lets_Encrypt_Staging",
		"Let's Encrypt Staging Server",
		"https://acme-staging-v02.api.letsencrypt.org/directory",
		1,
		strftime('%s', 'now'),
		strftime('%s', 'now')
);

-- insert default user/password (hash of default password isn't a secret)
INSERT INTO
		users (id, username, password_hash, created_at, updated_at)
	VALUES (
		1, -- this was a goof in the past, but it is what it is
		'admin',
		'$2a$12$q2zn2nvyGIGC1BfpORWS6.Y.Q1n8.R0.U9RtHn31m6WbaTqHiSjpG',
		strftime('%s', 'now'),
		strftime('%s', 'now')
);



-- +goose Down

DROP TABLE users;
DROP TABLE acme_orders;	
DROP TABLE certificates;
DROP TABLE acme_accounts;
DROP TABLE private_keys;
DROP TABLE acme_servers;
