package migrations_test

import (
	"certwarden-backend/pkg/storage/migrations"
	"context"
	"database/sql"
	"testing"
)

func insertDataV4(t *testing.T, db *sql.DB) {
	ctx, cancel := context.WithTimeout(context.Background(), migrations.MigrateDBTimeout)
	defer cancel()

	//
	// certificate -- with new attribute
	q := `
		INSERT INTO certificates
		VALUES
			(3, 3, 1, "some cert 3", "cert desc 3", "subj3.example.com",
				'["alt6","alt5"]', "org3", "ou3", "countr3", "state3", "city3", "apkey13", "apkey23",
				0, "some cmd3", '["val here3","another val3"]', 'aaabbbccc', 123453, 444443);
	`

	_, err := db.ExecContext(ctx, q)
	if err != nil {
		t.Fatalf("insertdatav4: failed to insert cert 3 (%s)", err)
	}
}

func validateDataV4(t *testing.T, db *sql.DB) {
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
			post_processing_command = '' AND post_processing_environment = '[]' AND post_processing_client_key = '';
	`

	rows := db.QueryRowContext(ctx, q)
	count := -1
	err := rows.Scan(&count)
	if err != nil {
		t.Fatalf("validatedatav4: failed to scan certificate 1 (%s)", err)
	}
	if count != 1 {
		t.Errorf("validatedatav4: failed to retrieve certificate 1 (row count expected 1 but got '%d')", count)
	}

	// new cert with new val
	q = `
		SELECT COUNT(*) FROM certificates
		WHERE id = 3 AND private_key_id = 3 AND acme_account_id = 1 AND name = 'some cert 3' AND description = 'cert desc 3' AND
			subject = 'subj3.example.com' AND subject_alts = '["alt6","alt5"]' AND
			csr_org = 'org3' AND csr_ou = 'ou3' AND csr_country = 'countr3' AND csr_state = 'state3' AND csr_city = 'city3' AND
			api_key = 'apkey13' AND api_key_new = 'apkey23' AND api_key_via_url = 0 AND created_at = 123453 AND updated_at = 444443 AND
			post_processing_command = 'some cmd3' AND post_processing_environment = '["val here3","another val3"]' AND
			post_processing_client_key = 'aaabbbccc';
	`

	rows = db.QueryRowContext(ctx, q)
	count = -1
	err = rows.Scan(&count)
	if err != nil {
		t.Fatalf("validatedatav4: failed to scan certificate 3 (%s)", err)
	}
	if count != 1 {
		t.Errorf("validatedatav4: failed to retrieve certificate 3 (row count expected 1 but got '%d')", count)
	}
}
