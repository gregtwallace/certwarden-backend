-- CHANGES v11 to v12:
-- - users:
--    - Add `COLLATE NOCASE` to `username` attribute (on the off chance there are duplicate
--      usernames this migration will fail and they will need to be manually fixed)
-- - orders:
--    - Add `COLLATE NOCASE` to `acme_location` (on the off chance there are duplicate
--      acme_location (url) this migration will fail and they will need to be manually 
--      fixed. This really shouldn't happen though.)
--    - Remove `acme_account_id` as it isn't actually used and can be traced through the
--      relevant `acme_account_id`.



-- +goose Up



-- rename old table
ALTER TABLE users RENAME TO users_old;
ALTER TABLE acme_orders RENAME TO acme_orders_old;


-- create new table
-- adds `COLLATE NOCASE` to `username`
CREATE TABLE users (
	id integer PRIMARY KEY AUTOINCREMENT NOT NULL UNIQUE,
	username text NOT NULL UNIQUE COLLATE NOCASE,
	password_hash NOT NULL,
	created_at integer NOT NULL,
	updated_at integer NOT NULL
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


-- copy data from old to new
INSERT
	INTO
		users (
    id, username, password_hash, created_at, updated_at
    )
	SELECT
		id, username, password_hash, created_at, updated_at
		FROM users_old;

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


-- drop old tables
DROP TABLE users_old;
DROP TABLE acme_orders_old;



-- +goose Down



-- rename old tables
ALTER TABLE users RENAME TO users_old;
ALTER TABLE acme_orders RENAME TO acme_orders_old;


-- create new tables (v10)
-- removes `COLLATE NOCASE` from `username`
CREATE TABLE users (
	id integer PRIMARY KEY AUTOINCREMENT NOT NULL UNIQUE,
	username text NOT NULL UNIQUE,
	password_hash NOT NULL,
	created_at integer NOT NULL,
	updated_at integer NOT NULL
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


-- copy data from old to new
INSERT
	INTO
		users (
    id, username, password_hash, created_at, updated_at
    )
	SELECT
		id, username, password_hash, created_at, updated_at
		FROM users_old;

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


-- drop old tables
DROP TABLE users_old;
DROP TABLE acme_orders_old;
