package migrations_test

import (
	"certwarden-backend/pkg/storage/migrations"
	"context"
	"database/sql"
	"testing"
)

func insertDataV10(t *testing.T, db *sql.DB) {
	ctx, cancel := context.WithTimeout(context.Background(), migrations.MigrateDBTimeout)
	defer cancel()

	// certificate -- with new attribute
	q := `
		INSERT INTO certificates
		VALUES
			(8, 8, 0, "cert8", "cert desc 8", "subj8.example.com",
				'["alt8","alt18"]', "org8", "ou8", "countr8", "state8", "city8", "[some csr obj array8]", 'a root8',
				"apkey18", "apkey28",	0, "some cmd8", '["val here8","another val8"]', 'example.com/client/8',
				'aaabbbccc8', 'profile x8', 55555558, 123458, 444448);
	`

	_, err := db.ExecContext(ctx, q)
	if err != nil {
		t.Fatalf("insertdatav10: failed to insert certificate 7 (%s)", err)
	}

	// order -- with new attribute
	q = `
		INSERT INTO acme_orders
		VALUES
			(2, 1, 2, "https://example.com/ord/12322", "valid2", 0, 'some err obj2', 123412,
				'["alt112","alt212"]', '["auth112","auth212"]', "example.com/final/12312", 2, "certurl2",
				"pem data here2", 12312, 34512, 'some root 2', 'profile x2', 11112, 22212);
	`

	_, err = db.ExecContext(ctx, q)
	if err != nil {
		t.Fatalf("insertdatav10: failed to insert order 2 (%s)", err)
	}
}

func validateDataV10(t *testing.T, db *sql.DB) {
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
			csr_extra_extensions = '[]' AND preferred_root_cn = '' AND last_access = 0 AND post_processing_client_address = ''
			AND profile = '';
	`

	rows := db.QueryRowContext(ctx, q)
	count := -1
	err := rows.Scan(&count)
	if err != nil {
		t.Fatalf("validatedatav10: failed to scan certificate 1 (%s)", err)
	}
	if count != 1 {
		t.Errorf("validatedatav10: failed to retrieve certificate 1 (row count expected 1 but got '%d')", count)
	}

	// new cert with new val
	q = `
		SELECT COUNT(*) FROM certificates
		WHERE id = 8 AND private_key_id = 8 AND acme_account_id = 0 AND name = 'cert8' AND description = 'cert desc 8' AND
			subject = 'subj8.example.com' AND subject_alts = '["alt8","alt18"]' AND
			csr_org = 'org8' AND csr_ou = 'ou8' AND csr_country = 'countr8' AND csr_state = 'state8' AND csr_city = 'city8' AND
			api_key = 'apkey18' AND api_key_new = 'apkey28' AND api_key_via_url = 0 AND created_at = 123458 AND updated_at = 444448 AND
			post_processing_command = 'some cmd8' AND post_processing_environment = '["val here8","another val8"]' AND
			post_processing_client_key = 'aaabbbccc8' AND csr_extra_extensions = '[some csr obj array8]' AND
			preferred_root_cn = 'a root8' AND last_access = 55555558 AND post_processing_client_address = 'example.com/client/8'
			AND profile = 'profile x8';
	`

	rows = db.QueryRowContext(ctx, q)
	count = -1
	err = rows.Scan(&count)
	if err != nil {
		t.Fatalf("validatedatav10: failed to scan certificate 8 (%s)", err)
	}
	if count != 1 {
		t.Errorf("validatedatav10: failed to retrieve certificate 8 (row count expected 1 but got '%d')", count)
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
			created_at = 111 AND updated_at = 222 AND chain_root_cn IS NULL AND profile IS NULL;
	`

	rows = db.QueryRowContext(ctx, q)
	count = -1
	err = rows.Scan(&count)
	if err != nil {
		t.Fatalf("validatedatav10: failed to scan acme order 0 (%s)", err)
	}
	if count != 1 {
		t.Errorf("validatedatav10: failed to retrieve acme order 0 (row count expected 1 but got '%d')", count)
	}

	// new order with new val
	q = `
		SELECT COUNT(*) FROM acme_orders
		WHERE id = 2 AND acme_account_id = 1 AND certificate_id = 2 AND acme_location = 'https://example.com/ord/12322' AND 
			status = 'valid2' AND known_revoked = 0 AND error = 'some err obj2' AND expires = 123412 AND
			dns_identifiers = '["alt112","alt212"]' AND authorizations = '["auth112","auth212"]' AND
			finalize = 'example.com/final/12312' AND finalized_key_id = 2 AND
			certificate_url = 'certurl2' AND pem = 'pem data here2' AND valid_from = 12312 AND valid_to = 34512 AND
			created_at = 11112 AND updated_at = 22212 AND chain_root_cn = 'some root 2' AND profile = 'profile x2';
	`

	rows = db.QueryRowContext(ctx, q)
	count = -1
	err = rows.Scan(&count)
	if err != nil {
		t.Fatalf("validatedatav10: failed to scan acme order 2 (%s)", err)
	}
	if count != 1 {
		t.Errorf("validatedatav10: failed to retrieve acme order 2 (row count expected 1 but got '%d')", count)
	}
}
