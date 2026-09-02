package migrations_test

import (
	"certwarden-backend/pkg/helpers_test"
	"certwarden-backend/pkg/storage/migrations"
	"context"
	"database/sql"
	"testing"
)

func insertDataV2(t *testing.T, db *sql.DB) {
	ctx, cancel := context.WithTimeout(context.Background(), migrations.MigrateDBTimeout)
	defer cancel()

	//
	// certificate -- without challenge_method
	q := `
		INSERT INTO certificates
		VALUES
			(1, 1, 1, "some cert 1", "cert desc 1", "subj1.example.com",
				"alt2, alt3", "org1", "ou1", "countr1", "state1", "city1", "apkey11", "apkey21",
				0, 123451, 444441);
	`

	_, err := db.ExecContext(ctx, q)
	if err != nil {
		t.Fatalf("insertdatav2: failed to insert cert 1 (%s)", err)
	}
}

func validateDataV2(t *testing.T, db *sql.DB, cameDownFromNextVer bool) {
	ctx, cancel := context.WithTimeout(context.Background(), migrations.MigrateDBTimeout)
	defer cancel()

	//
	// certificates
	// without the challenge_method attribute
	q := `
		SELECT COUNT(*) FROM certificates
		WHERE id = 1 AND private_key_id = 1 AND acme_account_id = 1 AND name = 'some cert 1' AND description = 'cert desc 1' AND 
			subject = 'subj1.example.com' AND subject_alts = 'alt2, alt3' AND
			csr_org = 'org1' AND csr_ou = 'ou1' AND csr_country = 'countr1' AND csr_state = 'state1' AND csr_city = 'city1' AND 
			api_key = 'apkey11' AND api_key_new = 'apkey21' AND api_key_via_url = 0 AND created_at = 123451 AND updated_at = 444441;
	`

	rows := db.QueryRowContext(ctx, q)
	count := -1
	err := rows.Scan(&count)
	if err != nil {
		t.Fatalf("validatedatav2: failed to scan certificate 1 (%s)", err)
	}
	if count != 1 {
		t.Errorf("validatedatav2: failed to retrieve certificate 1 (row count expected 1 but got '%d')", count)
	}

	// with challenge_method attribute (verify error)
	// certificates
	q = `
		SELECT COUNT(*) FROM certificates
		WHERE id = 1 AND private_key_id = 1 AND acme_account_id = 1 AND name = 'some cert 1' AND description = 'cert desc 1' AND 
			challenge_method = 'irrelevant' AND subject = 'subj1.example.com' AND subject_alts = 'alt2, alt3' AND
			csr_org = 'org1' AND csr_ou = 'ou1' AND csr_country = 'countr1' AND csr_state = 'state1' AND csr_city = 'city1' AND 
			api_key = 'apkey11' AND api_key_new = 'apkey21' AND api_key_via_url = 0 AND created_at = 123451 AND updated_at = 444441;
	`

	_, err = db.ExecContext(ctx, q)
	expectErr := helpers_test.NewTestErrorStringComp("no such column: challenge_method")
	if !helpers_test.ErrorsIs(err, expectErr) {
		t.Errorf("validatedatav0: expected err '%s' but got '%s'", expectErr, err)
	}
}
