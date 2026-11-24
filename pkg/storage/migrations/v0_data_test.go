package migrations_test

import (
	"certwarden-backend/pkg/helpers_test"
	"certwarden-backend/pkg/storage/migrations"
	"context"
	"database/sql"
	"testing"
)

// validateDataV0 confirms none of the app's tables exist
func validateDataV0(t *testing.T, db *sql.DB) {
	ctx, cancel := context.WithTimeout(context.Background(), migrations.MigrateDBTimeout)
	defer cancel()

	// check if `acme_servers` exists
	q := `
		SELECT * 
		FROM acme_servers
		LIMIT 1;
	`
	_, err := db.ExecContext(ctx, q)
	expectErr := helpers_test.NewTestErrorStringComp("no such table: acme_servers")
	if !helpers_test.ErrorsIs(err, expectErr) {
		t.Errorf("validatedatav0: expected err '%s' but got '%s'", expectErr, err)
	}

	// check if `acme_orders` exists
	q = `
		SELECT * 
		FROM acme_orders
		LIMIT 1;
	`
	_, err = db.ExecContext(ctx, q)
	expectErr = helpers_test.NewTestErrorStringComp("no such table: acme_orders")
	if !helpers_test.ErrorsIs(err, expectErr) {
		t.Errorf("validatedatav0: expected err '%s' but got '%s'", expectErr, err)
	}

	// check if `certificates` exists
	q = `
		SELECT * 
		FROM certificates
		LIMIT 1;
	`
	_, err = db.ExecContext(ctx, q)
	expectErr = helpers_test.NewTestErrorStringComp("no such table: certificates")
	if !helpers_test.ErrorsIs(err, expectErr) {
		t.Errorf("validatedatav0: expected err '%s' but got '%s'", expectErr, err)
	}

	// check if `acme_accounts` exists
	q = `
		SELECT * 
		FROM acme_accounts
		LIMIT 1;
	`
	_, err = db.ExecContext(ctx, q)
	expectErr = helpers_test.NewTestErrorStringComp("no such table: acme_accounts")
	if !helpers_test.ErrorsIs(err, expectErr) {
		t.Errorf("validatedatav0: expected err '%s' but got '%s'", expectErr, err)
	}

	// check if `private_keys` exists
	q = `
		SELECT * 
		FROM private_keys
		LIMIT 1;
	`
	_, err = db.ExecContext(ctx, q)
	expectErr = helpers_test.NewTestErrorStringComp("no such table: private_keys")
	if !helpers_test.ErrorsIs(err, expectErr) {
		t.Errorf("validatedatav0: expected err '%s' but got '%s'", expectErr, err)
	}

	// check if `users` exists
	q = `
		SELECT * 
		FROM users
		LIMIT 1;
	`
	_, err = db.ExecContext(ctx, q)
	expectErr = helpers_test.NewTestErrorStringComp("no such table: users")
	if !helpers_test.ErrorsIs(err, expectErr) {
		t.Errorf("validatedatav0: expected err '%s' but got '%s'", expectErr, err)
	}
}
