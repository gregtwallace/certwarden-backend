package migrations_test

import (
	"certwarden-backend/pkg/storage/migrations"
	"context"
	"database/sql"
	"testing"
)

func insertDataV5(t *testing.T, db *sql.DB) {
	ctx, cancel := context.WithTimeout(context.Background(), migrations.MigrateDBTimeout)
	defer cancel()

	//
	// certificate -- with new attribute
	q := `
		INSERT INTO certificates
		VALUES
			(4, 4, 1, "some cert 4", "cert desc 4", "subj4.example.com",
				'["alt4","alt5"]', "org4", "ou4", "countr4", "state4", "city4", "[some csr obj array4]", "apkey14", "apkey24",
				0, "some cmd4", '["val here4","another val4"]', 'aaabbbccc4', 123454, 444444);
	`

	_, err := db.ExecContext(ctx, q)
	if err != nil {
		t.Fatalf("insertdatav5: failed to insert cert 4 (%s)", err)
	}
}

func validateDataV5(t *testing.T, db *sql.DB) {
	ctx, cancel := context.WithTimeout(context.Background(), migrations.MigrateDBTimeout)
	defer cancel()

	//
	// certificates
	// new attribute has default val
	q := `
		SELECT COUNT(*) FROM certificates
		WHERE id = 1 AND private_key_id = 1 AND acme_account_id = 1 AND name = 'some cert 1' AND description = 'cert desc 1' AND 
			subject = 'subj1.example.com' AND subject_alts = '["alt2","alt3"]' AND
			csr_org = 'org1' AND csr_ou = 'ou1' AND csr_country = 'countr1' AND csr_state = 'state1' AND csr_city = 'city1' AND 
			api_key = 'apkey11' AND api_key_new = 'apkey21' AND api_key_via_url = 0 AND created_at = 123451 AND updated_at = 444441 AND
			post_processing_command = '' AND post_processing_environment = '[]' AND post_processing_client_key = '' AND
			csr_extra_extensions = '[]';
	`

	rows := db.QueryRowContext(ctx, q)
	count := -1
	err := rows.Scan(&count)
	if err != nil {
		t.Fatalf("validatedatav5: failed to scan certificate 1 (%s)", err)
	}
	if count != 1 {
		t.Errorf("validatedatav5: failed to retrieve certificate 1 (row count expected 1 but got '%d')", count)
	}

	// new cert with new val
	q = `
		SELECT COUNT(*) FROM certificates
		WHERE id = 4 AND private_key_id = 4 AND acme_account_id = 1 AND name = 'some cert 4' AND description = 'cert desc 4' AND
			subject = 'subj4.example.com' AND subject_alts = '["alt4","alt5"]' AND
			csr_org = 'org4' AND csr_ou = 'ou4' AND csr_country = 'countr4' AND csr_state = 'state4' AND csr_city = 'city4' AND
			api_key = 'apkey14' AND api_key_new = 'apkey24' AND api_key_via_url = 0 AND created_at = 123454 AND updated_at = 444444 AND
			post_processing_command = 'some cmd4' AND post_processing_environment = '["val here4","another val4"]' AND
			post_processing_client_key = 'aaabbbccc4' AND csr_extra_extensions = '[some csr obj array4]';
	`

	rows = db.QueryRowContext(ctx, q)
	count = -1
	err = rows.Scan(&count)
	if err != nil {
		t.Fatalf("validatedatav5: failed to scan certificate 4 (%s)", err)
	}
	if count != 1 {
		t.Errorf("validatedatav5: failed to retrieve certificate 4 (row count expected 1 but got '%d')", count)
	}
}
