package migrations_test

import (
	"certwarden-backend/pkg/storage/migrations"
	"context"
	"database/sql"
	"testing"
)

func insertDataV11(t *testing.T, db *sql.DB) {
	ctx, cancel := context.WithTimeout(context.Background(), migrations.MigrateDBTimeout)
	defer cancel()

	// order -- with new attribute
	q := `
		INSERT INTO acme_orders
		VALUES
			(3, 1, 2, "https://example.com/ord/123223", "valid3", 0, 'some err obj3', 1234123,
				'["alt1123","alt2123"]', '["auth1123","auth2123"]', "example.com/final/123123", 3, "certurl3",
				"pem data here3", 12313, 34513, 'some root 3', 'profile x3', '{an ari object 3}', 11113, 22213);
	`

	_, err := db.ExecContext(ctx, q)
	if err != nil {
		t.Fatalf("insertdatav11: failed to insert order 3 (%s)", err)
	}
}

func validateDataV11(t *testing.T, db *sql.DB) {
	ctx, cancel := context.WithTimeout(context.Background(), migrations.MigrateDBTimeout)
	defer cancel()

	//
	// order
	// new attribute has default val
	q := `
		SELECT COUNT(*) FROM acme_orders
		WHERE id = 0 AND acme_account_id = 0 AND certificate_id = 0 AND acme_location = 'https://example.com/ord/123' AND 
			status = 'invalid' AND known_revoked = 0 AND error IS null AND expires = 1234 AND dns_identifiers = '["alt1","alt2"]' AND
			authorizations = '["auth1","auth2"]' AND finalize = 'example.com/final/123' AND finalized_key_id = 0 AND
			certificate_url = 'certurl' AND pem = 'pem data here' AND valid_from = 123 AND valid_to = 345 AND
			created_at = 111 AND updated_at = 222 AND chain_root_cn IS NULL AND profile IS NULL AND renewal_info IS NULL;
	`

	rows := db.QueryRowContext(ctx, q)
	count := -1
	err := rows.Scan(&count)
	if err != nil {
		t.Fatalf("validatedatav11: failed to scan acme order 0 (%s)", err)
	}
	if count != 1 {
		t.Errorf("validatedatav11: failed to retrieve acme order 0 (row count expected 1 but got '%d')", count)
	}

	// new order with new val
	q = `
		SELECT COUNT(*) FROM acme_orders
		WHERE id = 3 AND acme_account_id = 1 AND certificate_id = 2 AND acme_location = 'https://example.com/ord/123223' AND 
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
		t.Fatalf("validatedatav11: failed to scan acme order 3 (%s)", err)
	}
	if count != 1 {
		t.Errorf("validatedatav11: failed to retrieve acme order 3 (row count expected 3 but got '%d')", count)
	}

	// orders v12: to ensure v12 "Down" is working correctly
	// this should insert without error (same location as order 3)
	// NOTE: this is under validate, not insert, since it isn't actually a net change to the db records
	q = `
		INSERT INTO acme_orders
		VALUES
			(1123, 0, 2, "https://EXAMPLE.com/ord/123223", "valid3", 0, 'some err obj3', 1234123,
				'["alt1123","alt2123"]', '["auth1123","auth2123"]', "example.com/final/123123", 3, "certurl3",
				"pem data here3", 12313, 34513, 'some root 3', 'profile x3', '{an ari object 3}', 11113, 22213);
	`

	_, err = db.ExecContext(ctx, q)
	if err != nil {
		t.Fatalf("insertdatav11: failed to insert order 1123 (%s)", err)
	}

	// but then delete it after confirming no error, so "Up" to 12 works correctly
	q = `
		DELETE FROM acme_orders
		WHERE id = 1123;
	`

	_, err = db.ExecContext(ctx, q)
	if err != nil {
		t.Fatalf("insertdatav11: failed to delete order 1123 (%s)", err)
	}

	// users v12: to ensure v12 "Down" is working correctly
	// this should insert without error (same username as user 1)
	// NOTE: this is under validate, not insert, since it isn't actually a net change to the db records
	q = `
		INSERT INTO
				users (id, username, password_hash, created_at, updated_at)
			VALUES (
				222,
				'aDMIn',
				'xxyyzz',
				987,
				654
		);
	`

	_, err = db.ExecContext(ctx, q)
	if err != nil {
		t.Fatalf("insertdatav11: failed to insert user 222 (%s)", err)
	}

	// but then delete it after confirming no error, so "Up" to 12 works correctly
	q = `
		DELETE FROM users
		WHERE id = 222;
	`

	_, err = db.ExecContext(ctx, q)
	if err != nil {
		t.Fatalf("insertdatav11: failed to delete user 222 (%s)", err)
	}
}
