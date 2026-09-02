-- CHANGES v2 to v3:
-- - certificates:
--     - Add `post_processing_command` attribute
--		 - Add `post_processing_environment` attribute
--		 - Modify 'subject_alts' from comma separated strings to valid json array object
-- - acme_orders:
--		 - Modify `dns_identifiers` from comma separated strings to valid json array object
--		 - Modify `authorizations` from comma separated strings to valid json array object


-- +goose Up


-- rename old table
-- NOTE: orders is also copied due to the way FK from orders links to certificates
ALTER TABLE acme_orders RENAME TO acme_orders_old;
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
  post_processing_command text NOT NULL DEFAULT "",
  post_processing_environment text NOT NULL DEFAULT "[]",
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

-- copy data from old to new
INSERT INTO certificates
  SELECT id, private_key_id, acme_account_id, name, description, subject, subject_alts, csr_org, csr_ou,
    csr_country, csr_state, csr_city, api_key, api_key_new, api_key_via_url, '',
    '[]', created_at, updated_at
  FROM certificates_old;

INSERT INTO acme_orders SELECT * FROM acme_orders_old;

-- modify data in `cerificates.subject_alts`
UPDATE certificates
  SET subject_alts = case when subject_alts is "" then "[]" else '["' || replace(subject_alts, ',', '","') || '"]' end;

-- modify data in `acme_orders.dns_identifiers` and `acme_orders.authorizations`
UPDATE acme_orders
  SET
    dns_identifiers = case when dns_identifiers is "" then "[]" else '["' || replace(dns_identifiers, ',', '","') || '"]' end,
    authorizations = case when authorizations is "" then "[]" else '["' || replace(authorizations, ',', '","') || '"]' end;

-- drop old tables
DROP TABLE acme_orders_old;
DROP TABLE certificates_old;



-- +goose Down


-- rename old tables
ALTER TABLE acme_orders RENAME TO acme_orders_old;
ALTER TABLE certificates RENAME TO certificates_old;

-- create new tables (v2)
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
INSERT INTO certificates
  SELECT id, private_key_id, acme_account_id, name, description, subject, subject_alts, csr_org, csr_ou,
    csr_country, csr_state, csr_city, api_key, api_key_new, api_key_via_url, created_at, updated_at
  FROM certificates_old;

INSERT INTO acme_orders SELECT * FROM acme_orders_old;

-- modify data in `cerificates.subject_alts`
UPDATE certificates
  SET subject_alts = replace(replace(replace(subject_alts, '[', ''), ']', ''), '"', '');

-- modify data in `acme_orders.dns_identifiers` and `acme_orders.authorizations`
UPDATE acme_orders
  SET
    dns_identifiers = replace(replace(replace(dns_identifiers, '[', ''), ']', ''), '"', ''),
    authorizations = replace(replace(replace(authorizations, '[', ''), ']', ''), '"', '');

-- drop old tables
DROP TABLE acme_orders_old;
DROP TABLE certificates_old;
