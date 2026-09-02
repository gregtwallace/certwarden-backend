-- CHANGES v7 to v8:
-- - certificates:
--     - Add 'last_access' attribute
-- - private_keys:
--     - Add 'last_access' attribute


-- +goose Up


-- rename old tables
-- NOTE: orders & acme_accounts are also copied due to the way FKs link
ALTER TABLE acme_orders RENAME TO acme_orders_old;
ALTER TABLE certificates RENAME TO certificates_old;
ALTER TABLE acme_accounts RENAME TO acme_accounts_old;
ALTER TABLE private_keys RENAME TO private_keys_old;

-- create new tables
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
	last_access integer NOT NULL DEFAULT 0,
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
	subject text NOT NULL,
	subject_alts text NOT NULL,
	csr_org text NOT NULL,
	csr_ou text NOT NULL,
	csr_country text NOT NULL,
	csr_state text NOT NULL,
	csr_city text NOT NULL,
	csr_extra_extensions text NOT NULL DEFAULT "[]",
	preferred_root_cn text NOT NULL DEFAULT "",
	api_key text NOT NULL,
	api_key_new text NOT NULL DEFAULT '',
	api_key_via_url integer NOT NULL DEFAULT 0 CHECK(api_key_via_url IN (0,1)),
	post_processing_command text NOT NULL DEFAULT "",
	post_processing_environment text NOT NULL DEFAULT "[]",
	post_processing_client_key text NOT NULL DEFAULT "",
	last_access integer NOT NULL DEFAULT 0,
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
  chain_root_cn text,
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


-- copy data from old to new
INSERT INTO private_keys
  SELECT id, name, description, algorithm, pem, api_key, api_key_new, api_key_disabled, api_key_via_url, 0,
		created_at, updated_at
  FROM private_keys_old;

INSERT INTO acme_accounts SELECT * FROM acme_accounts_old;

INSERT INTO certificates
  SELECT id, private_key_id, acme_account_id, name, description, subject, subject_alts, csr_org, csr_ou,
    csr_country, csr_state, csr_city, csr_extra_extensions, preferred_root_cn, api_key, api_key_new, 
		api_key_via_url, post_processing_command, post_processing_environment, post_processing_client_key,
		0, created_at, updated_at
  FROM certificates_old;

INSERT INTO acme_orders SELECT * FROM acme_orders_old;

-- drop old tables
DROP TABLE acme_orders_old;
DROP TABLE certificates_old;
DROP TABLE acme_accounts_old;
DROP TABLE private_keys_old;



-- +goose Down


-- rename old tables
ALTER TABLE acme_orders RENAME TO acme_orders_old;
ALTER TABLE certificates RENAME TO certificates_old;
ALTER TABLE acme_accounts RENAME TO acme_accounts_old;
ALTER TABLE private_keys RENAME TO private_keys_old;

-- create new tables (v7)
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
  subject text NOT NULL,
  subject_alts text NOT NULL,
  csr_org text NOT NULL,
  csr_ou text NOT NULL,
  csr_country text NOT NULL,
  csr_state text NOT NULL,
  csr_city text NOT NULL,
  csr_extra_extensions text NOT NULL DEFAULT "[]",
  preferred_root_cn text NOT NULL DEFAULT "",
  api_key text NOT NULL,
  api_key_new text NOT NULL DEFAULT '',
  api_key_via_url integer NOT NULL DEFAULT 0 CHECK(api_key_via_url IN (0,1)),
  post_processing_command text NOT NULL DEFAULT "",
  post_processing_environment text NOT NULL DEFAULT "[]",
  post_processing_client_key text NOT NULL DEFAULT "",
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
  chain_root_cn text,
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

-- copy data from old to new
INSERT INTO private_keys
  SELECT id, name, description, algorithm, pem, api_key, api_key_new, api_key_disabled, api_key_via_url, 
		created_at, updated_at
  FROM private_keys_old;

INSERT INTO acme_accounts SELECT * FROM acme_accounts_old;

INSERT INTO certificates
  SELECT id, private_key_id, acme_account_id, name, description, subject, subject_alts, csr_org, csr_ou,
    csr_country, csr_state, csr_city, csr_extra_extensions, preferred_root_cn, api_key, api_key_new, 
		api_key_via_url, post_processing_command, post_processing_environment, post_processing_client_key,
		created_at, updated_at
  FROM certificates_old;

INSERT INTO acme_orders SELECT * FROM acme_orders_old;

-- drop old tables
DROP TABLE acme_orders_old;
DROP TABLE certificates_old;
DROP TABLE acme_accounts_old;
DROP TABLE private_keys_old;
