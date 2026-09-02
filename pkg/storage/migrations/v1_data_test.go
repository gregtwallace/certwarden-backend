package migrations_test

import (
	"certwarden-backend/pkg/storage/migrations"
	"context"
	"database/sql"
	"testing"
)

func insertDataV1(t *testing.T, db *sql.DB) {
	ctx, cancel := context.WithTimeout(context.Background(), migrations.MigrateDBTimeout)
	defer cancel()

	//
	// acme_servers -- skip -- should be inserted by goose during v0 -> v1

	//
	// private_keys
	q := `
		INSERT INTO private_keys
		VALUES
			(0, "key0", "some desc", "an-alg", "pemdata", "apikey123",
				"apikeynew234", 1, 1, 123, 456 ),
			(1, "key1", "some desc 1", "an-alg-1", "pemdata1", "apikey1231",
				"apikeynew2341", 0, 0, 111, 333 ),
			(2, "key2", "some desc 2", "an-alg-2", "pemdata2", "apikey1232",
				"apikeynew2342", 1, 0, 222, 444 ),
			(3, "key3", "some desc 3", "an-alg-3", "pemdata3", "apikey1233",
				"apikeynew2343", 1, 0, 2223, 4443 ),
			(4, "key4", "some desc 4", "an-alg-4", "pemdata4", "apikey1234",
				"apikeynew2344", 1, 0, 2224, 4444 ),
			(5, "key5", "some desc 5", "an-alg-5", "pemdata5", "apikey1235",
				"apikeynew2345", 1, 0, 2225, 4445 ),
			(7, "key7", "some desc 7", "an-alg-7", "pemdata7", "apikey1237",
				"apikeynew2347", 1, 0, 2227, 4447 ),
			(8, "key8", "some desc 8", "an-alg-8", "pemdata8", "apikey1238",
				"apikeynew2348", 1, 0, 2228, 4448 );
	`

	_, err := db.ExecContext(ctx, q)
	if err != nil {
		t.Fatalf("insertdatav1: failed to insert private key 0 (%s)", err)
	}

	//
	// acme_accounts
	q = `
		INSERT INTO acme_accounts
		VALUES
			(0, "acct0", 0, "some acct", "invalid", "contact@example.com",
				1, 555, 1222, "example.com/acct123", 0 ),
			(1, "acct1", 1, "some acct 1", "valid", "contact2@example.com",
				0, 777, 888, "example.com/acct1232", 1 );
	`

	_, err = db.ExecContext(ctx, q)
	if err != nil {
		t.Fatalf("insertdatav1: failed to insert acme server 0 (%s)", err)
	}

	//
	// certificates
	q = `
		INSERT INTO certificates
		VALUES
			(0, 0, 0, "some cert", "cert desc", "chall-method-1", "subj.example.com",
				"alt1,alt2", "org", "ou", "countr", "state", "city", "apkey1", "apkey2",
				1, 12345, 44444);
	`

	_, err = db.ExecContext(ctx, q)
	if err != nil {
		t.Fatalf("insertdatav1: failed to insert cert 0 (%s)", err)
	}

	//
	// acme_orders
	q = `
		INSERT INTO acme_orders
		VALUES
			(0, 0, 0, "https://example.com/ord/123", "invalid", 0, null, 1234,
				"alt1,alt2", "auth1,auth2", "example.com/final/123", 0, "certurl",
				"pem data here", 123, 345, 111, 222);
	`

	_, err = db.ExecContext(ctx, q)
	if err != nil {
		t.Fatalf("insertdatav1: failed to insert order 0 (%s)", err)
	}

	//
	// users -- skip -- should be inserted by goose during v0 -> v1
}

func validateDataV1(t *testing.T, db *sql.DB, cameDownFromNextVer bool) {
	ctx, cancel := context.WithTimeout(context.Background(), migrations.MigrateDBTimeout)
	defer cancel()

	//
	// acme_servers
	q := `
		SELECT COUNT(*) FROM acme_servers
		WHERE id = 0 AND name = 'Lets_Encrypt' AND description = 'Let''s Encrypt Production Server' AND 
			directory_url = 'https://acme-v02.api.letsencrypt.org/directory' AND is_staging = 0 AND
			created_at > 0 AND updated_at > 0;
	`

	rows := db.QueryRowContext(ctx, q)
	count := -1
	err := rows.Scan(&count)
	if err != nil {
		t.Fatalf("validatedatav1: failed to scan acme server 0 (%s)", err)
	}
	if count != 1 {
		t.Errorf("validatedatav1: failed to retrieve acme server 0 (row count expected 1 but got '%d')", count)
	}

	q = `
		SELECT COUNT(*) FROM acme_servers
		WHERE id = 1 AND name = 'Lets_Encrypt_Staging' AND description = 'Let''s Encrypt Staging Server' AND 
			directory_url = 'https://acme-staging-v02.api.letsencrypt.org/directory' AND is_staging = 1 AND
			created_at > 0 AND updated_at > 0;
	`

	rows = db.QueryRowContext(ctx, q)
	count = -1
	err = rows.Scan(&count)
	if err != nil {
		t.Fatalf("validatedatav1: failed to scan acme server 1 (%s)", err)
	}
	if count != 1 {
		t.Errorf("validatedatav1: failed to retrieve acme server 1 (row count expected 1 but got '%d')", count)
	}

	//
	// private_keys
	q = `
		SELECT COUNT(*) FROM private_keys
		WHERE id = 0 AND name = 'key0' AND description = 'some desc' AND algorithm = 'an-alg' AND pem = 'pemdata' AND
			api_key = 'apikey123' AND api_key_new = 'apikeynew234' AND api_key_disabled = 1 AND
			api_key_via_url = 1 AND created_at = 123 AND updated_at = 456;
	`

	rows = db.QueryRowContext(ctx, q)
	count = -1
	err = rows.Scan(&count)
	if err != nil {
		t.Fatalf("validatedatav1: failed to scan private key 0 (%s)", err)
	}
	if count != 1 {
		t.Errorf("validatedatav1: failed to retrieve private key 0 (row count expected 1 but got '%d')", count)
	}

	//
	// acme_accounts
	q = `
		SELECT COUNT(*) FROM acme_accounts
		WHERE id = 0 AND name = 'acct0' AND private_key_id = 0 AND description = 'some acct' AND 
			status = 'invalid' AND email = 'contact@example.com' AND accepted_tos = 1 AND
			created_at = 555 AND updated_at = 1222 AND kid = 'example.com/acct123' AND acme_server_id = 0;
	`

	rows = db.QueryRowContext(ctx, q)
	count = -1
	err = rows.Scan(&count)
	if err != nil {
		t.Fatalf("validatedatav1: failed to scan acme account 0 (%s)", err)
	}
	if count != 1 {
		t.Errorf("validatedatav1: failed to retrieve acme account 0 (row count expected 1 but got '%d')", count)
	}

	//
	// certificates
	q = `
		SELECT COUNT(*) FROM certificates
		WHERE id = 0 AND private_key_id = 0 AND acme_account_id = 0 AND name = 'some cert' AND description = 'cert desc' AND 
			challenge_method = $1 AND subject = 'subj.example.com' AND subject_alts = 'alt1,alt2' AND
			csr_org = 'org' AND csr_ou = 'ou' AND csr_country = 'countr' AND csr_state = 'state' AND csr_city = 'city' AND 
			api_key = 'apkey1' AND api_key_new = 'apkey2' AND api_key_via_url = 1 AND created_at = 12345 AND updated_at = 44444;
	`
	challVal := "chall-method-1"
	if cameDownFromNextVer {
		challVal = ""
	}

	rows = db.QueryRowContext(ctx, q,
		challVal,
	)
	count = -1
	err = rows.Scan(&count)
	if err != nil {
		t.Fatalf("validatedatav1: failed to scan certificate 0 (%s)", err)
	}
	if count != 1 {
		t.Errorf("validatedatav1: failed to retrieve certificate 0 (row count expected 1 but got '%d')", count)
	}

	//
	// acme_orders
	q = `
		SELECT COUNT(*) FROM acme_orders
		WHERE id = 0 AND acme_account_id = 0 AND certificate_id = 0 AND acme_location = 'https://example.com/ord/123' AND 
			status = 'invalid' AND known_revoked = 0 AND error IS null AND expires = 1234 AND dns_identifiers = 'alt1,alt2' AND
			authorizations = 'auth1,auth2' AND finalize = 'example.com/final/123' AND finalized_key_id = 0 AND
			certificate_url = 'certurl' AND pem = 'pem data here' AND valid_from = 123 AND valid_to = 345 AND
			created_at = 111 AND updated_at = 222;
	`

	rows = db.QueryRowContext(ctx, q)
	count = -1
	err = rows.Scan(&count)
	if err != nil {
		t.Fatalf("validatedatav1: failed to scan acme order 0 (%s)", err)
	}
	if count != 1 {
		t.Errorf("validatedatav1: failed to retrieve acme order 0 (row count expected 1 but got '%d')", count)
	}

	//
	// users
	q = `
		SELECT COUNT(*) FROM users
		WHERE id = 1 AND username = 'admin' AND created_at > 0 AND updated_at > 0 AND
			password_hash = '$2a$12$q2zn2nvyGIGC1BfpORWS6.Y.Q1n8.R0.U9RtHn31m6WbaTqHiSjpG';
	`

	rows = db.QueryRowContext(ctx, q)
	count = -1
	err = rows.Scan(&count)
	if err != nil {
		t.Fatalf("validatedatav1: failed to scan user 1 (%s)", err)
	}
	if count != 1 {
		t.Errorf("validatedatav1: failed to retrieve user 1 (row count expected 1 but got '%d')", count)
	}
}
