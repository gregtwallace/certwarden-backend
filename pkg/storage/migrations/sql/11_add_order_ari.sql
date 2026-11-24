-- CHANGES v10 to v11:
-- - orders:
-- 	 - Add 'renewal_info' attribute



-- +goose Up



-- rename old table
ALTER TABLE acme_orders RENAME TO acme_orders_old;


-- create new table
-- adds `renewal_info`
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


-- copy data from old to new
INSERT
	INTO
		acme_orders (
    id, acme_account_id, certificate_id, acme_location, status, known_revoked, error, expires,
    dns_identifiers, authorizations, finalize, finalized_key_id, certificate_url, pem, valid_from, valid_to,
    chain_root_cn, profile, renewal_info, created_at, updated_at
    )
	SELECT
		id, acme_account_id, certificate_id, acme_location, status, known_revoked, error, expires, 
		dns_identifiers, authorizations, finalize, finalized_key_id, certificate_url, pem, valid_from, valid_to,
		chain_root_cn, profile, null, created_at, updated_at
		FROM acme_orders_old;


-- drop old tables
DROP TABLE acme_orders_old;



-- +goose Down



-- rename old tables
ALTER TABLE acme_orders RENAME TO acme_orders_old;


-- create new tables (v10)
-- drops `renewal_info`
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
		acme_orders (
    id, acme_account_id, certificate_id, acme_location, status, known_revoked, error, expires,
    dns_identifiers, authorizations, finalize, finalized_key_id, certificate_url, pem, valid_from, valid_to,
    chain_root_cn, profile,created_at, updated_at
    )
	SELECT
		id, acme_account_id, certificate_id, acme_location, status, known_revoked, error, expires, 
		dns_identifiers, authorizations, finalize, finalized_key_id, certificate_url, pem, valid_from, valid_to,
		chain_root_cn, profile, created_at, updated_at
		FROM acme_orders_old;


-- drop old tables
DROP TABLE acme_orders_old;
