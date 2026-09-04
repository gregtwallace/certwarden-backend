-- CHANGES v8 to v9:
-- - certificates:
--     - Add 'post_processing_client_address' field/column



-- +goose Up



-- rename old tables
-- NOTE: orders is also copied due to the way FK from orders links to certificates
ALTER TABLE acme_orders RENAME TO acme_orders_old;
ALTER TABLE certificates RENAME TO certificates_old;


-- create new tables
-- adds `post_processing_client_address`
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

-- no modifications
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
INSERT
	INTO
		certificates (
    id, private_key_id, acme_account_id, name, description, subject, subject_alts, csr_org, csr_ou,
    csr_country, csr_state, csr_city, csr_extra_extensions, preferred_root_cn, api_key, api_key_new,
    api_key_via_url, post_processing_command, post_processing_environment, post_processing_client_address,
    post_processing_client_key, last_access, created_at, updated_at
    )
	SELECT
    id, private_key_id, acme_account_id, name, description, subject, subject_alts, csr_org, csr_ou,
    csr_country, csr_state, csr_city, csr_extra_extensions, preferred_root_cn, api_key, api_key_new,
    api_key_via_url, post_processing_command, post_processing_environment, '',
    post_processing_client_key, last_access, created_at, updated_at
		FROM certificates_old;

-- no changes
INSERT
	INTO
		acme_orders (
    id, acme_account_id, certificate_id, acme_location, status, known_revoked, error, expires,
    dns_identifiers, authorizations, finalize, finalized_key_id, certificate_url, pem, valid_from, valid_to,
    chain_root_cn, created_at, updated_at
    )
	SELECT
		id, acme_account_id, certificate_id, acme_location, status, known_revoked, error, expires, 
		dns_identifiers, authorizations, finalize, finalized_key_id, certificate_url, pem, valid_from, valid_to,
		chain_root_cn, created_at, updated_at
		FROM acme_orders_old;


-- populate `post_processing_client_address` if `post_processing_client_key` is not empty and 'subject'
-- is not a wildcard (i.e. does not start with '*'); previously app used the `subject` for this value,
-- so just copy it
UPDATE certificates
  SET
    post_processing_client_address = subject
  WHERE
    LENGTH(post_processing_client_key) > 0
    AND
    subject NOT LIKE '*%';


-- drop old tables
DROP TABLE acme_orders_old;
DROP TABLE certificates_old;



-- +goose Down



-- rename old tables
ALTER TABLE acme_orders RENAME TO acme_orders_old;
ALTER TABLE certificates RENAME TO certificates_old;


-- create new tables (v8)
-- drops `post_processing_client_address`
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

-- no modification
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
INSERT
	INTO
		certificates (
    id, private_key_id, acme_account_id, name, description, subject, subject_alts, csr_org, csr_ou,
    csr_country, csr_state, csr_city, csr_extra_extensions, preferred_root_cn, api_key, api_key_new,
    api_key_via_url, post_processing_command, post_processing_environment, post_processing_client_key,
    last_access, created_at, updated_at
    )
	SELECT
		id, private_key_id, acme_account_id, name, description, subject, subject_alts, csr_org, csr_ou,
    csr_country, csr_state, csr_city, csr_extra_extensions, preferred_root_cn, api_key, api_key_new,
    api_key_via_url, post_processing_command, post_processing_environment, post_processing_client_key,
    last_access, created_at, updated_at
		FROM certificates_old;

-- no changes
INSERT
	INTO
		acme_orders (
    id, acme_account_id, certificate_id, acme_location, status, known_revoked, error, expires,
    dns_identifiers, authorizations, finalize, finalized_key_id, certificate_url, pem, valid_from, valid_to,
    chain_root_cn, created_at, updated_at
    )
	SELECT
		id, acme_account_id, certificate_id, acme_location, status, known_revoked, error, expires, 
		dns_identifiers, authorizations, finalize, finalized_key_id, certificate_url, pem, valid_from, valid_to,
		chain_root_cn, created_at, updated_at
		FROM acme_orders_old;


-- no need to reverse population step above (attribute is dropped)


-- drop old tables
DROP TABLE acme_orders_old;
DROP TABLE certificates_old;
