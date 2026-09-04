package migrations_test

import (
	"certwarden-backend/pkg/storage/migrations"
	"context"
	"database/sql"
	"testing"
)

func insertDataV3(t *testing.T, db *sql.DB) {
	ctx, cancel := context.WithTimeout(context.Background(), migrations.MigrateDBTimeout)
	defer cancel()

	//
	// certificate -- with new fields
	q := `
		INSERT INTO certificates
		VALUES
			(2, 2, 1, "some cert 2", "cert desc 2", "subj2.example.com",
				'["alt4","alt5"]', "org2", "ou2", "countr2", "state2", "city2", "apkey12", "apkey22",
				1, "some cmd", '["val here","another val"]', 123452, 444442);
	`

	_, err := db.ExecContext(ctx, q)
	if err != nil {
		t.Fatalf("insertdatav3: failed to insert cert 2 (%s)", err)
	}
}

func validateDataV3(t *testing.T, db *sql.DB) {
	ctx, cancel := context.WithTimeout(context.Background(), migrations.MigrateDBTimeout)
	defer cancel()

	//
	// certificates
	// with post_processing_command & post_processing_environment which should both be default val
	// and converted val for subject_alts
	q := `
		SELECT COUNT(*) FROM certificates
		WHERE id = 1 AND private_key_id = 1 AND acme_account_id = 1 AND name = 'some cert 1' AND description = 'cert desc 1' AND 
			subject = 'subj1.example.com' AND subject_alts = '["alt2","alt3"]' AND
			csr_org = 'org1' AND csr_ou = 'ou1' AND csr_country = 'countr1' AND csr_state = 'state1' AND csr_city = 'city1' AND 
			api_key = 'apkey11' AND api_key_new = 'apkey21' AND api_key_via_url = 0 AND created_at = 123451 AND updated_at = 444441 AND
			post_processing_command = '' AND post_processing_environment = '[]';
	`

	rows := db.QueryRowContext(ctx, q)
	count := -1
	err := rows.Scan(&count)
	if err != nil {
		t.Fatalf("validatedatav3: failed to scan certificate 1 (%s)", err)
	}
	if count != 1 {
		t.Errorf("validatedatav3: failed to retrieve certificate 1 (row count expected 1 but got '%d')", count)
	}

	q = `
		SELECT COUNT(*) FROM certificates
		WHERE id = 2 AND private_key_id = 2 AND acme_account_id = 1 AND name = 'some cert 2' AND description = 'cert desc 2' AND 
			subject = 'subj2.example.com' AND subject_alts = '["alt4","alt5"]' AND
			csr_org = 'org2' AND csr_ou = 'ou2' AND csr_country = 'countr2' AND csr_state = 'state2' AND csr_city = 'city2' AND 
			api_key = 'apkey12' AND api_key_new = 'apkey22' AND api_key_via_url = 1 AND created_at = 123452 AND updated_at = 444442 AND
			post_processing_command = 'some cmd' AND post_processing_environment = '["val here","another val"]';
	`

	rows = db.QueryRowContext(ctx, q)
	count = -1
	err = rows.Scan(&count)
	if err != nil {
		t.Fatalf("validatedatav3: failed to scan certificate 2 (%s)", err)
	}
	if count != 1 {
		t.Errorf("validatedatav3: failed to retrieve certificate 2 (row count expected 1 but got '%d')", count)
	}

	//
	// acme_orders
	// with modified dns_identifiers and authorizations values
	q = `
		SELECT COUNT(*) FROM acme_orders
		WHERE id = 0 AND acme_account_id = 0 AND certificate_id = 0 AND acme_location = 'https://example.com/ord/123' AND 
			status = 'invalid' AND known_revoked = 0 AND error IS null AND expires = 1234 AND dns_identifiers = '["alt1","alt2"]' AND
			authorizations = '["auth1","auth2"]' AND finalize = 'example.com/final/123' AND finalized_key_id = 0 AND
			certificate_url = 'certurl' AND pem = 'pem data here' AND valid_from = 123 AND valid_to = 345 AND
			created_at = 111 AND updated_at = 222;
	`

	rows = db.QueryRowContext(ctx, q)
	count = -1
	err = rows.Scan(&count)
	if err != nil {
		t.Fatalf("validatedatav3: failed to scan acme order 0 (%s)", err)
	}
	if count != 1 {
		t.Errorf("validatedatav3: failed to retrieve acme order 0 (row count expected 1 but got '%d')", count)
	}
}
