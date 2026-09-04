-- CHANGES v11 to v12:
-- - private_keys:
--    - Allow `last_access` to be NULL (and convert 0 values to NULL)
-- - certificates:
--    - Allow `last_access` to be NULL (and convert 0 values to NULL)
-- - acme_orders:
--    - Add `COLLATE NOCASE` to `acme_location` (on the off chance there are duplicate
--      acme_location (url) this migration will fail and they will need to be manually 
--      fixed. This really shouldn't happen though.)
--    - Remove `acme_account_id` as it isn't actually used and can be traced through the
--      relevant `acme_account_id`.
-- - users:
--    - Add `COLLATE NOCASE` to `username` attribute (on the off chance there are duplicate
--      usernames this migration will fail and they will need to be manually fixed)



-- +goose Up



-- rename old tables
-- NOTE: acme_accounts are also copied due to the way FKs link
ALTER TABLE users RENAME TO users_old;
ALTER TABLE acme_orders RENAME TO acme_orders_old;
ALTER TABLE certificates RENAME TO certificates_old;
ALTER TABLE acme_accounts RENAME TO acme_accounts_old;
ALTER TABLE private_keys RENAME TO private_keys_old;


-- create new tables
-- adds `last_access`
-- remove `NOT NULL` from `last_access` and change default to `DEFAULT NULL`
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
	last_access integer DEFAULT NULL,
	created_at integer NOT NULL,
	updated_at integer NOT NULL
);

-- no modifications
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

-- remove `NOT NULL` from `last_access` and change default to `DEFAULT NULL`
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
	post_processing_client_address text NOT NULL DEFAULT "",
	post_processing_client_key text NOT NULL DEFAULT "",
  profile text NOT NULL DEFAULT "",
	last_access integer DEFAULT NULL,
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

-- adds `COLLATE NOCASE` to `acme_location`
-- removes `acme_account_id` and related FK
CREATE TABLE acme_orders (
  id integer PRIMARY KEY AUTOINCREMENT NOT NULL UNIQUE,
  certificate_id integer NOT NULL,
  acme_location text NOT NULL UNIQUE COLLATE NOCASE,
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
  profile text DEFAULT NULL,
  renewal_info text DEFAULT NULL,
  created_at integer NOT NULL,
  updated_at integer NOT NULL,
  FOREIGN KEY (finalized_key_id)
    REFERENCES private_keys (id)
      ON DELETE SET NULL
      ON UPDATE NO ACTION,
  FOREIGN KEY (certificate_id)
    REFERENCES certificates (id)
      ON DELETE CASCADE
      ON UPDATE NO ACTION
);

-- adds `COLLATE NOCASE` to `username`
CREATE TABLE users (
	id integer PRIMARY KEY AUTOINCREMENT NOT NULL UNIQUE,
	username text NOT NULL UNIQUE COLLATE NOCASE,
	password_hash NOT NULL,
	created_at integer NOT NULL,
	updated_at integer NOT NULL
);


-- copy data from old to new
-- no changes
INSERT
	INTO
		private_keys (
    id, name, description, algorithm, pem, api_key, api_key_new, api_key_disabled, api_key_via_url, 
    last_access, created_at, updated_at
    )
	SELECT
		id, name, description, algorithm, pem, api_key, api_key_new, api_key_disabled, api_key_via_url,
    last_access, created_at, updated_at
		FROM private_keys_old;

-- no changes
INSERT
	INTO
		acme_accounts (
    id, name, private_key_id, description, status, email, accepted_tos, created_at,
    updated_at, kid, acme_server_id
    )
	SELECT
		id, name, private_key_id, description, status, email, accepted_tos, created_at,
    updated_at, kid, acme_server_id
		FROM acme_accounts_old;

-- no changes
INSERT
	INTO
		certificates (
    id, private_key_id, acme_account_id, name, description, subject, subject_alts, csr_org, csr_ou,
    csr_country, csr_state, csr_city, csr_extra_extensions, preferred_root_cn, api_key, api_key_new,
    api_key_via_url, post_processing_command, post_processing_environment, post_processing_client_address,
    post_processing_client_key, profile, last_access, created_at, updated_at
    )
	SELECT
    id, private_key_id, acme_account_id, name, description, subject, subject_alts, csr_org, csr_ou,
    csr_country, csr_state, csr_city, csr_extra_extensions, preferred_root_cn, api_key, api_key_new,
    api_key_via_url, post_processing_command, post_processing_environment, post_processing_client_address,
    post_processing_client_key,	profile, last_access, created_at, updated_at
		FROM certificates_old;

-- omits `acme_account_id`
INSERT
	INTO
		acme_orders (
    id, certificate_id, acme_location, status, known_revoked, error, expires,
    dns_identifiers, authorizations, finalize, finalized_key_id, certificate_url, pem, valid_from, valid_to,
    chain_root_cn, profile, renewal_info, created_at, updated_at
    )
	SELECT
		id, certificate_id, acme_location, status, known_revoked, error, expires, 
		dns_identifiers, authorizations, finalize, finalized_key_id, certificate_url, pem, valid_from, valid_to,
		chain_root_cn, profile, renewal_info, created_at, updated_at
		FROM acme_orders_old;

INSERT
	INTO
		users (
    id, username, password_hash, created_at, updated_at
    )
	SELECT
		id, username, password_hash, created_at, updated_at
		FROM users_old;


-- modify `last_access` to convert `0` to `NULL`
UPDATE private_keys
  SET last_access = NULL
  WHERE last_access = 0;

UPDATE certificates
  SET last_access = NULL
  WHERE last_access = 0;


-- drop old tables
DROP TABLE users_old;
DROP TABLE acme_orders_old;
DROP TABLE certificates_old;
DROP TABLE acme_accounts_old;
DROP TABLE private_keys_old;



-- +goose Down



-- rename old tables
ALTER TABLE users RENAME TO users_old;
ALTER TABLE acme_orders RENAME TO acme_orders_old;
ALTER TABLE certificates RENAME TO certificates_old;
ALTER TABLE acme_accounts RENAME TO acme_accounts_old;
ALTER TABLE private_keys RENAME TO private_keys_old;


-- create new tables (v11)
-- changes `last_access` back to `NOT NULL DEFAULT 0`
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

-- no modifications
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

-- changes `last_access` back to `NOT NULL DEFAULT 0`
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
	post_processing_client_address text NOT NULL DEFAULT "",
	post_processing_client_key text NOT NULL DEFAULT "",
  profile text NOT NULL DEFAULT "",
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

-- removes `COLLATE NOCASE` from `acme_location`
-- add back `acme_account_id`
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
  profile text DEFAULT NULL,
  renewal_info text DEFAULT NULL,
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

-- removes `COLLATE NOCASE` from `username`
CREATE TABLE users (
	id integer PRIMARY KEY AUTOINCREMENT NOT NULL UNIQUE,
	username text NOT NULL UNIQUE,
	password_hash NOT NULL,
	created_at integer NOT NULL,
	updated_at integer NOT NULL
);


-- copy data from old to new
-- change `last_access` NULL to 0
-- do in one step due to the NOT NULL
INSERT
	INTO
		private_keys (
    id, name, description, algorithm, pem, api_key, api_key_new, api_key_disabled, api_key_via_url, 
    last_access,
    created_at, updated_at
    )
	SELECT
		id, name, description, algorithm, pem, api_key, api_key_new, api_key_disabled, api_key_via_url,
    COALESCE(last_access, 0),
    created_at, updated_at
		FROM private_keys_old;

-- no changes
INSERT
	INTO
		acme_accounts (
    id, name, private_key_id, description, status, email, accepted_tos, created_at,
    updated_at, kid, acme_server_id
    )
	SELECT
		id, name, private_key_id, description, status, email, accepted_tos, created_at,
    updated_at, kid, acme_server_id
		FROM acme_accounts_old;

-- change `last_access` NULL to 0
-- do in one step due to the NOT NULL
INSERT
	INTO
		certificates (
    id, private_key_id, acme_account_id, name, description, subject, subject_alts, csr_org, csr_ou,
    csr_country, csr_state, csr_city, csr_extra_extensions, preferred_root_cn, api_key, api_key_new,
    api_key_via_url, post_processing_command, post_processing_environment, post_processing_client_address,
    post_processing_client_key, profile,
    last_access,
    created_at, updated_at
    )
	SELECT
    id, private_key_id, acme_account_id, name, description, subject, subject_alts, csr_org, csr_ou,
    csr_country, csr_state, csr_city, csr_extra_extensions, preferred_root_cn, api_key, api_key_new,
    api_key_via_url, post_processing_command, post_processing_environment, post_processing_client_address,
    post_processing_client_key,	profile,
    COALESCE(last_access, 0),
    created_at, updated_at
		FROM certificates_old;

-- use certificate_id->certificates->acme_account_id to reconstruct acme_account_id
-- do in one step due to the NOT NULL and FK requirement
INSERT
	INTO
		acme_orders (
    id, acme_account_id, certificate_id, acme_location, status, known_revoked, error, expires,
    dns_identifiers, authorizations, finalize, finalized_key_id, certificate_url, pem, valid_from, valid_to,
    chain_root_cn, profile, renewal_info, created_at, updated_at
    )
	SELECT
		aoo.id, c.acme_account_id, aoo.certificate_id, aoo.acme_location, aoo.status, aoo.known_revoked, aoo.error,
    aoo.expires, aoo.dns_identifiers, aoo.authorizations, aoo.finalize, aoo.finalized_key_id, aoo.certificate_url,
    aoo.pem, aoo.valid_from, aoo.valid_to, aoo.chain_root_cn, aoo.profile, aoo.renewal_info, aoo.created_at, aoo.updated_at
    FROM
		  acme_orders_old aoo
		  LEFT JOIN certificates c on (aoo.certificate_id = c.id);

INSERT
	INTO
		users (
    id, username, password_hash, created_at, updated_at
    )
	SELECT
		id, username, password_hash, created_at, updated_at
		FROM users_old;


-- drop old tables
DROP TABLE users_old;
DROP TABLE acme_orders_old;
DROP TABLE certificates_old;
DROP TABLE acme_accounts_old;
DROP TABLE private_keys_old;
