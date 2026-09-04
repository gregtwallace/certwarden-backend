-- WARNING: goose down cannot recreate the data that was dropped during up. As such,
-- the resulting attribute will be blank and must be manually fixed.

-- CHANGES v1 to v2:
-- - certificates:
--     - Delete 'challenge_method' attribute



-- +goose Up



-- rename old tables
-- NOTE: orders is also copied due to the way FK from orders links to certificates
ALTER TABLE acme_orders RENAME TO acme_orders_old;
ALTER TABLE certificates RENAME TO certificates_old;


-- create new tables
-- remove `challenge_method`
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

-- no changes
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


-- copy data from old to new
-- `challenge_method` omitted
INSERT
	INTO
		certificates (id, private_key_id, acme_account_id, name, description, subject, subject_alts, csr_org, csr_ou,
    csr_country, csr_state, csr_city, api_key, api_key_new, api_key_via_url, created_at, updated_at)
	SELECT
		id, private_key_id, acme_account_id, name, description, subject, subject_alts, csr_org, csr_ou,
    csr_country, csr_state, csr_city, api_key, api_key_new, api_key_via_url, created_at, updated_at
		FROM certificates_old;

-- no modifications
INSERT
	INTO
		acme_orders (id, acme_account_id, certificate_id, acme_location, status, known_revoked, error,
		expires, dns_identifiers, authorizations, finalize, finalized_key_id, certificate_url, pem, valid_from,
		valid_to, created_at, updated_at)
	SELECT
		id, acme_account_id, certificate_id, acme_location, status, known_revoked, error, expires, 
		dns_identifiers, authorizations, finalize, finalized_key_id, certificate_url, pem, valid_from, valid_to,
		created_at, updated_at
		FROM acme_orders_old;


-- drop old tables
DROP TABLE acme_orders_old;
DROP TABLE certificates_old;



-- +goose Down



-- rename 'old' table
ALTER TABLE acme_orders RENAME TO acme_orders_old;
ALTER TABLE certificates RENAME TO certificates_old;


-- create new table (v1)
-- add `challenge_method`
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

-- no changes
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


-- copy data from old to new
-- add back `challenge_method` with empty value
INSERT
	INTO
		certificates (id, private_key_id, acme_account_id, name, description, challenge_method, subject, subject_alts, csr_org, csr_ou,
    csr_country, csr_state, csr_city, api_key, api_key_new, api_key_via_url, created_at, updated_at)
	SELECT
		id, private_key_id, acme_account_id, name, description, '', subject, subject_alts, csr_org, csr_ou,
    csr_country, csr_state, csr_city, api_key, api_key_new, api_key_via_url, created_at, updated_at
  	FROM certificates_old;

-- no modifications
INSERT
	INTO
		acme_orders (id, acme_account_id, certificate_id, acme_location, status, known_revoked, error,
		expires, dns_identifiers, authorizations, finalize, finalized_key_id, certificate_url, pem, valid_from,
		valid_to, created_at, updated_at)
	SELECT
		id, acme_account_id, certificate_id, acme_location, status, known_revoked, error, expires, 
		dns_identifiers, authorizations, finalize, finalized_key_id, certificate_url, pem, valid_from, valid_to,
		created_at, updated_at
		FROM acme_orders_old;


-- drop old tables
DROP TABLE acme_orders_old;
DROP TABLE certificates_old;
