package migrations_test

import (
	"certwarden-backend/pkg/storage/migrations"
	"context"
	"database/sql"
	"testing"
)

func insertDataV7(t *testing.T, db *sql.DB) {
	ctx, cancel := context.WithTimeout(context.Background(), migrations.MigrateDBTimeout)
	defer cancel()

	//
	// certificate -- with new attribute
	q := `
		INSERT INTO certificates
		VALUES
			(5, 5, 1, "cert5", "cert desc 5", "*.subj5.example.com",
				'["alt5","alt15"]', "org5", "ou5", "countr5", "state5", "city5", "[some csr obj array5]", 'a root',
				"apkey15", "apkey25",	0, "some cmd5", '["val here5","another val5"]', 'aaabbbccc5', 123455, 444445);
	`

	_, err := db.ExecContext(ctx, q)
	if err != nil {
		t.Fatalf("insertdatav7: failed to insert cert 5 (%s)", err)
	}

	//
	// order -- with new attribute
	q = `
		INSERT INTO acme_orders
		VALUES
			(1, 1, 1, "https://example.com/ord/1231", "invalid1", 1, 'some err obj', 12341,
				'["alt11","alt21"]', '["auth11","auth21"]', "example.com/final/1231", 1, "certurl1",
				"pem data here1", 1231, 3451, 'some root 1', 1111, 2221);
	`

	_, err = db.ExecContext(ctx, q)
	if err != nil {
		t.Fatalf("insertdatav7: failed to insert order 1 (%s)", err)
	}
}

func validateDataV7(t *testing.T, db *sql.DB) {
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
			csr_extra_extensions = '[]' AND preferred_root_cn = '';
	`

	rows := db.QueryRowContext(ctx, q)
	count := -1
	err := rows.Scan(&count)
	if err != nil {
		t.Fatalf("validatedatav7: failed to scan certificate 1 (%s)", err)
	}
	if count != 1 {
		t.Errorf("validatedatav7: failed to retrieve certificate 1 (row count expected 1 but got '%d')", count)
	}

	// new cert with new val
	q = `
		SELECT COUNT(*) FROM certificates
		WHERE id = 5 AND private_key_id = 5 AND acme_account_id = 1 AND name = 'cert5' AND description = 'cert desc 5' AND
			subject = '*.subj5.example.com' AND subject_alts = '["alt5","alt15"]' AND
			csr_org = 'org5' AND csr_ou = 'ou5' AND csr_country = 'countr5' AND csr_state = 'state5' AND csr_city = 'city5' AND
			api_key = 'apkey15' AND api_key_new = 'apkey25' AND api_key_via_url = 0 AND created_at = 123455 AND updated_at = 444445 AND
			post_processing_command = 'some cmd5' AND post_processing_environment = '["val here5","another val5"]' AND
			post_processing_client_key = 'aaabbbccc5' AND csr_extra_extensions = '[some csr obj array5]' AND
			preferred_root_cn = 'a root';
	`

	rows = db.QueryRowContext(ctx, q)
	count = -1
	err = rows.Scan(&count)
	if err != nil {
		t.Fatalf("validatedatav7: failed to scan certificate 5 (%s)", err)
	}
	if count != 1 {
		t.Errorf("validatedatav7: failed to retrieve certificate 5 (row count expected 1 but got '%d')", count)
	}

	//
	// order
	// new attribute has default val
	q = `
		SELECT COUNT(*) FROM acme_orders
		WHERE id = 0 AND acme_account_id = 0 AND certificate_id = 0 AND acme_location = 'https://example.com/ord/123' AND 
			status = 'invalid' AND known_revoked = 0 AND error IS null AND expires = 1234 AND dns_identifiers = '["alt1","alt2"]' AND
			authorizations = '["auth1","auth2"]' AND finalize = 'example.com/final/123' AND finalized_key_id = 0 AND
			certificate_url = 'certurl' AND pem = 'pem data here' AND valid_from = 123 AND valid_to = 345 AND
			created_at = 111 AND updated_at = 222 AND chain_root_cn IS NULL;
	`

	rows = db.QueryRowContext(ctx, q)
	count = -1
	err = rows.Scan(&count)
	if err != nil {
		t.Fatalf("validatedatav7: failed to scan acme order 0 (%s)", err)
	}
	if count != 1 {
		t.Errorf("validatedatav7: failed to retrieve acme order 0 (row count expected 1 but got '%d')", count)
	}

	q = `
		SELECT COUNT(*) FROM acme_orders
		WHERE id = 1 AND acme_account_id = 1 AND certificate_id = 1 AND acme_location = 'https://example.com/ord/1231' AND 
			status = 'invalid1' AND known_revoked = 1 AND error = 'some err obj' AND expires = 12341 AND
			dns_identifiers = '["alt11","alt21"]' AND authorizations = '["auth11","auth21"]' AND
			finalize = 'example.com/final/1231' AND finalized_key_id = 1 AND
			certificate_url = 'certurl1' AND pem = 'pem data here1' AND valid_from = 1231 AND valid_to = 3451 AND
			created_at = 1111 AND updated_at = 2221 AND chain_root_cn = 'some root 1';
	`

	rows = db.QueryRowContext(ctx, q)
	count = -1
	err = rows.Scan(&count)
	if err != nil {
		t.Fatalf("validatedatav7: failed to scan acme order 1 (%s)", err)
	}
	if count != 1 {
		t.Errorf("validatedatav7: failed to retrieve acme order 1 (row count expected 1 but got '%d')", count)
	}
}
