package migrations_test

import (
	"certwarden-backend/pkg/helpers_test"
	"certwarden-backend/pkg/storage/migrations"
	"context"
	"database/sql"
	"testing"
)

func insertDataV12(t *testing.T, db *sql.DB) {
	ctx, cancel := context.WithTimeout(context.Background(), migrations.MigrateDBTimeout)
	defer cancel()

	// users
	// ensure insert of existing user username, different case, fails
	q := `
		INSERT INTO
				users (id, username, password_hash, created_at, updated_at)
			VALUES (
				123,
				'aDMIn',
				'irrelevant',
				1234,
				5678
		);
	`

	_, err := db.ExecContext(ctx, q)
	expectedErr := helpers_test.NewTestErrorStringComp("UNIQUE constraint failed: users.username")
	if !helpers_test.ErrorsIs(err, expectedErr) {
		t.Errorf("insertdatav12: insert user 123 err expected '%s' but got '%s'", helpers_test.ErrorToVal(expectedErr), helpers_test.ErrorToVal(err))
	}

	// orders
	// ensure insert of existing order acme_location, different case, fails
	q = `
		INSERT INTO acme_orders
		VALUES
			(1123, 2, "https://EXAMPLE.com/ord/123223", "valid3", 0, 'some err obj3', 1234123,
				'["alt1123","alt2123"]', '["auth1123","auth2123"]', "example.com/final/123123", 3, "certurl3",
				"pem data here3", 12313, 34513, 'some root 3', 'profile x3', '{an ari object 3}', 11113, 22213);
	`

	_, err = db.ExecContext(ctx, q)
	expectedErr = helpers_test.NewTestErrorStringComp("UNIQUE constraint failed: acme_orders.acme_location")
	if !helpers_test.ErrorsIs(err, expectedErr) {
		t.Errorf("insertdatav12: insert order 1123 err expected '%s' but got '%s'", helpers_test.ErrorToVal(expectedErr), helpers_test.ErrorToVal(err))
	}

	// private keys
	// ensure insert of NULL last_access works
	q = `
		INSERT INTO private_keys
		VALUES
			(9, "key9", "some desc 9", "an-alg-9", "pemdata9", "apikey1239",
				"apikeynew2349", 1, 1, NULL, 2229, 4449 );
	`

	_, err = db.ExecContext(ctx, q)
	if err != nil {
		t.Fatalf("insertdatav12: failed to insert private key 9 (%s)", err)
	}

	// certificate -- ensure insert of NULL last_access works
	q = `
		INSERT INTO certificates
		VALUES
			(9, 9, 0, "cert9", "cert desc 9", "subj9.example.com",
				'["alt9","alt19"]', "org9", "ou9", "countr9", "state9", "city9", "[some csr obj array9]", 'a root9',
				"apkey19", "apkey29",	1, "some cmd9", '["val here9","another val9"]', 'example.com/client/9',
				'aaabbbccc9', 'profile x9', NULL, 123459, 444449);
	`

	_, err = db.ExecContext(ctx, q)
	if err != nil {
		t.Fatalf("insertdatav12: failed to insert certificate 9 (%s)", err)
	}
}

func validateDataV12(t *testing.T, db *sql.DB) {
	ctx, cancel := context.WithTimeout(context.Background(), migrations.MigrateDBTimeout)
	defer cancel()

	// order
	// verify attribute is gone but record is fine
	q := `
		SELECT COUNT(*) FROM acme_orders
		WHERE id = 0 AND certificate_id = 0 AND acme_location = 'https://example.com/ord/123' AND 
			status = 'invalid' AND known_revoked = 0 AND error IS null AND expires = 1234 AND dns_identifiers = '["alt1","alt2"]' AND
			authorizations = '["auth1","auth2"]' AND finalize = 'example.com/final/123' AND finalized_key_id = 0 AND
			certificate_url = 'certurl' AND pem = 'pem data here' AND valid_from = 123 AND valid_to = 345 AND
			created_at = 111 AND updated_at = 222 AND chain_root_cn IS NULL AND profile IS NULL AND renewal_info IS NULL;
	`

	rows := db.QueryRowContext(ctx, q)
	count := -1
	err := rows.Scan(&count)
	if err != nil {
		t.Fatalf("validatedatav12: failed to scan acme order 0 (%s)", err)
	}
	if count != 1 {
		t.Errorf("validatedatav12: failed to retrieve acme order 0 (row count expected 1 but got '%d')", count)
	}

	q = `
		SELECT acme_account_id FROM acme_orders
		WHERE id = 0;
	`

	_, err = db.ExecContext(ctx, q)
	expectedErr := helpers_test.NewTestErrorStringComp("no such column: acme_account_id")
	if !helpers_test.ErrorsIs(err, expectedErr) {
		t.Errorf("validatedatav12: select 'acme_account_id' expected error '%s' but got '%s'", helpers_test.ErrorToVal(expectedErr), helpers_test.ErrorToVal(err))
	}

	//
	// users
	// ensure it fetches with wrong username case
	q = `
		SELECT COUNT(*) FROM users
		WHERE id = 1 AND username = 'aDmIN' AND created_at > 0 AND updated_at > 0 AND
			password_hash = '$2a$12$q2zn2nvyGIGC1BfpORWS6.Y.Q1n8.R0.U9RtHn31m6WbaTqHiSjpG';
	`

	rows = db.QueryRowContext(ctx, q)
	count = -1
	err = rows.Scan(&count)
	if err != nil {
		t.Fatalf("validatedatav12: failed to scan user 1 (%s)", err)
	}
	if count != 1 {
		t.Errorf("validatedatav12: failed to retrieve user 1 (row count expected 1 but got '%d')", count)
	}

	//
	// users
	// ensure it fetches with wrong acme_location case
	q = `
		SELECT COUNT(*) FROM acme_orders
		WHERE id = 3 AND certificate_id = 2 AND acme_location = 'https://EXAMPLE.com/ord/123223' AND 
			status = 'valid3' AND known_revoked = 0 AND error = 'some err obj3' AND expires = 1234123 AND
			dns_identifiers = '["alt1123","alt2123"]' AND authorizations = '["auth1123","auth2123"]' AND
			finalize = 'example.com/final/123123' AND finalized_key_id = 3 AND
			certificate_url = 'certurl3' AND pem = 'pem data here3' AND valid_from = 12313 AND valid_to = 34513 AND
			created_at = 11113 AND updated_at = 22213 AND chain_root_cn = 'some root 3' AND profile = 'profile x3' AND
			renewal_info = '{an ari object 3}';
	`

	rows = db.QueryRowContext(ctx, q)
	count = -1
	err = rows.Scan(&count)
	if err != nil {
		t.Fatalf("validatedatav12: failed to scan acme order 3 (%s)", err)
	}
	if count != 1 {
		t.Errorf("validatedatav12: failed to retrieve acme order 3 (row count expected 3 but got '%d')", count)
	}

	//
	// private keys
	// confirm inserted key is NULL
	q = `
		SELECT COUNT(*) FROM private_keys
		WHERE id = 9 AND name = 'key9' AND description = 'some desc 9' AND algorithm = 'an-alg-9' AND pem = 'pemdata9' AND
			api_key = 'apikey1239' AND api_key_new = 'apikeynew2349' AND api_key_disabled = 1 AND
			api_key_via_url = 1 AND last_access IS NULL AND created_at = 2229 AND updated_at = 4449;
	`

	rows = db.QueryRowContext(ctx, q)
	count = -1
	err = rows.Scan(&count)
	if err != nil {
		t.Fatalf("validatedatav12: failed to scan private key 9 (%s)", err)
	}
	if count != 1 {
		t.Errorf("validatedatav12: failed to retrieve private key 9 (row count expected 1 but got '%d')", count)
	}

	// confirm old 0 value key is now NULL
	q = `
		SELECT COUNT(*) FROM private_keys
		WHERE id = 4 AND name = 'key4' AND description = 'some desc 4' AND algorithm = 'an-alg-4' AND
			pem = 'pemdata4' AND api_key = 'apikey1234' AND api_key_new = 'apikeynew2344' AND api_key_disabled = 1 AND
			api_key_via_url = 0 AND created_at = 2224 AND updated_at = 4444 AND last_access IS NULL;
	`

	rows = db.QueryRowContext(ctx, q)
	count = -1
	err = rows.Scan(&count)
	if err != nil {
		t.Fatalf("validatedatav12: failed to scan private key 4 (%s)", err)
	}
	if count != 1 {
		t.Errorf("validatedatav12: failed to retrieve private key 4 (row count expected 1 but got '%d')", count)
	}

	//
	// certificates
	// confirm inserted cert is NULL
	q = `
		SELECT COUNT(*) FROM certificates
		WHERE id = 9 AND private_key_id = 9 AND acme_account_id = 0 AND name = 'cert9' AND description = 'cert desc 9' AND
			subject = 'subj9.example.com' AND subject_alts = '["alt9","alt19"]' AND
			csr_org = 'org9' AND csr_ou = 'ou9' AND csr_country = 'countr9' AND csr_state = 'state9' AND csr_city = 'city9' AND
			api_key = 'apkey19' AND api_key_new = 'apkey29' AND api_key_via_url = 1 AND created_at = 123459 AND updated_at = 444449 AND
			post_processing_command = 'some cmd9' AND post_processing_environment = '["val here9","another val9"]' AND
			post_processing_client_key = 'aaabbbccc9' AND csr_extra_extensions = '[some csr obj array9]' AND
			preferred_root_cn = 'a root9' AND last_access IS NULL AND post_processing_client_address = 'example.com/client/9'
			AND profile = 'profile x9';
	`

	rows = db.QueryRowContext(ctx, q)
	count = -1
	err = rows.Scan(&count)
	if err != nil {
		t.Fatalf("validatedatav12: failed to scan certificate 9 (%s)", err)
	}
	if count != 1 {
		t.Errorf("validatedatav12: failed to retrieve certificate 9 (row count expected 1 but got '%d')", count)
	}

	// confirm old 0 value cert is now NULL
	q = `
		SELECT COUNT(*) FROM certificates
		WHERE id = 1 AND private_key_id = 1 AND acme_account_id = 1 AND name = 'some cert 1' AND description = 'cert desc 1' AND
			subject = 'subj1.example.com' AND subject_alts = '["alt2","alt3"]' AND
			csr_org = 'org1' AND csr_ou = 'ou1' AND csr_country = 'countr1' AND csr_state = 'state1' AND csr_city = 'city1' AND
			api_key = 'apkey11' AND api_key_new = 'apkey21' AND api_key_via_url = 0 AND created_at = 123451 AND updated_at = 444441 AND
			post_processing_command = '' AND post_processing_environment = '[]' AND post_processing_client_key = '' AND
			csr_extra_extensions = '[]' AND preferred_root_cn = '' AND last_access IS NULL AND post_processing_client_address = ''
			AND profile = '';
	`

	rows = db.QueryRowContext(ctx, q)
	count = -1
	err = rows.Scan(&count)
	if err != nil {
		t.Fatalf("validatedatav12: failed to scan certificate 1 (%s)", err)
	}
	if count != 1 {
		t.Errorf("validatedatav12: failed to retrieve certificate 1 (row count expected 1 but got '%d')", count)
	}
}
