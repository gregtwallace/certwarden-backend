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
}
