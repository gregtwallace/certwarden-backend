-- WARNING: goose down cannot recreate the data that was dropped during up. As such,
-- the resulting attribute will be blank and must be manually fixed.

-- CHANGES v1 to v2:
-- - certificates:
--     - Delete 'challenge_method' attribute


-- +goose Up

-- rename old table
ALTER TABLE certificates RENAME TO certificates_old;

-- create new table
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

-- copy data from old to new
INSERT INTO certificates
  SELECT id, private_key_id, acme_account_id, name, description, subject, subject_alts, csr_org, csr_ou,
    csr_country, csr_state, csr_city, api_key, api_key_new, api_key_via_url, created_at, updated_at
  FROM certificates_old;

-- drop old table
DROP TABLE certificates_old;



-- +goose Down


-- rename 'old' table
ALTER TABLE certificates RENAME TO certificates_old;

-- create new table (v1)
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

-- copy data from old to new
INSERT INTO certificates
  SELECT id, private_key_id, acme_account_id, name, description, '', subject, subject_alts, csr_org, csr_ou,
    csr_country, csr_state, csr_city, api_key, api_key_new, api_key_via_url, created_at, updated_at
  FROM certificates_old;

-- drop old table
DROP TABLE certificates_old;
