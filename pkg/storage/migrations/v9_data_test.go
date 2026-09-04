package migrations_test

import (
	"certwarden-backend/pkg/storage/migrations"
	"context"
	"database/sql"
	"testing"
)

func insertDataV9(t *testing.T, db *sql.DB) {
	ctx, cancel := context.WithTimeout(context.Background(), migrations.MigrateDBTimeout)
	defer cancel()

	//
	// certificate -- with new attribute
	q := `
		INSERT INTO certificates
		VALUES
			(7, 7, 1, "cert7", "cert desc 7", "subj7.example.com",
				'["alt7","alt17"]', "org7", "ou7", "countr7", "state7", "city7", "[some csr obj array7]", 'a root7',
				"apkey17", "apkey27",	0, "some cmd7", '["val here7","another val7"]', 'example.com/client/', 
				'aaabbbccc7', 55555557, 123457, 444447);
	`

	_, err := db.ExecContext(ctx, q)
	if err != nil {
		t.Fatalf("insertdatav9: failed to insert certificate 7 (%s)", err)
	}
}

func validateDataV9(t *testing.T, db *sql.DB) {
	ctx, cancel := context.WithTimeout(context.Background(), migrations.MigrateDBTimeout)
	defer cancel()

	//
	// certificates
	// new attribute has default val (and post_processing_client_key = '')
	q := `
		SELECT COUNT(*) FROM certificates
		WHERE id = 1 AND private_key_id = 1 AND acme_account_id = 1 AND name = 'some cert 1' AND description = 'cert desc 1' AND 
			subject = 'subj1.example.com' AND subject_alts = '["alt2","alt3"]' AND
			csr_org = 'org1' AND csr_ou = 'ou1' AND csr_country = 'countr1' AND csr_state = 'state1' AND csr_city = 'city1' AND 
			api_key = 'apkey11' AND api_key_new = 'apkey21' AND api_key_via_url = 0 AND created_at = 123451 AND updated_at = 444441 AND
			post_processing_command = '' AND post_processing_environment = '[]' AND post_processing_client_key = '' AND
			csr_extra_extensions = '[]' AND preferred_root_cn = '' AND last_access = 0 AND post_processing_client_address = '';
	`

	rows := db.QueryRowContext(ctx, q)
	count := -1
	err := rows.Scan(&count)
	if err != nil {
		t.Fatalf("validatedatav9: failed to scan certificate 1 (%s)", err)
	}
	if count != 1 {
		t.Errorf("validatedatav9: failed to retrieve certificate 1 (row count expected 1 but got '%d')", count)
	}

	// new attribute has default val (and post_processing_client_key = '<some val>')
	q = `
		SELECT COUNT(*) FROM certificates
		WHERE id = 6 AND private_key_id = 6 AND acme_account_id = 0 AND name = 'cert6' AND description = 'cert desc 6' AND
			subject = 'subj6.example.com' AND subject_alts = '["alt6","alt16"]' AND
			csr_org = 'org6' AND csr_ou = 'ou6' AND csr_country = 'countr6' AND csr_state = 'state6' AND csr_city = 'city6' AND
			api_key = 'apkey16' AND api_key_new = 'apkey26' AND api_key_via_url = 0 AND created_at = 123456 AND updated_at = 444446 AND
			post_processing_command = 'some cmd6' AND post_processing_environment = '["val here6","another val6"]' AND
			post_processing_client_key = 'aaabbbccc6' AND csr_extra_extensions = '[some csr obj array6]' AND
			preferred_root_cn = 'a root6' AND last_access = 55555555 AND post_processing_client_address = 'subj6.example.com';
	`

	rows = db.QueryRowContext(ctx, q)
	count = -1
	err = rows.Scan(&count)
	if err != nil {
		t.Fatalf("validatedatav9: failed to scan certificate 6 (%s)", err)
	}
	if count != 1 {
		t.Errorf("validatedatav9: failed to retrieve certificate 6 (row count expected 1 but got '%d')", count)
	}

	// new attribute has default val (and post_processing_client_key = '<some val>', but subject is wildcard (*%))
	q = `
		SELECT COUNT(*) FROM certificates
		WHERE id = 5 AND private_key_id = 5 AND acme_account_id = 1 AND name = 'cert5' AND description = 'cert desc 5' AND
			subject = '*.subj5.example.com' AND subject_alts = '["alt5","alt15"]' AND
			csr_org = 'org5' AND csr_ou = 'ou5' AND csr_country = 'countr5' AND csr_state = 'state5' AND csr_city = 'city5' AND
			api_key = 'apkey15' AND api_key_new = 'apkey25' AND api_key_via_url = 0 AND created_at = 123455 AND updated_at = 444445 AND
			post_processing_command = 'some cmd5' AND post_processing_environment = '["val here5","another val5"]' AND
			post_processing_client_key = 'aaabbbccc5' AND csr_extra_extensions = '[some csr obj array5]' AND
			preferred_root_cn = 'a root' AND post_processing_client_address = '';
	`

	rows = db.QueryRowContext(ctx, q)
	count = -1
	err = rows.Scan(&count)
	if err != nil {
		t.Fatalf("validatedatav9: failed to scan certificate 5 (%s)", err)
	}
	if count != 1 {
		t.Errorf("validatedatav9: failed to retrieve certificate 5 (row count expected 1 but got '%d')", count)
	}

	// new cert with new val
	q = `
		SELECT COUNT(*) FROM certificates
		WHERE id = 7 AND private_key_id = 7 AND acme_account_id = 1 AND name = 'cert7' AND description = 'cert desc 7' AND
			subject = 'subj7.example.com' AND subject_alts = '["alt7","alt17"]' AND
			csr_org = 'org7' AND csr_ou = 'ou7' AND csr_country = 'countr7' AND csr_state = 'state7' AND csr_city = 'city7' AND
			api_key = 'apkey17' AND api_key_new = 'apkey27' AND api_key_via_url = 0 AND created_at = 123457 AND updated_at = 444447 AND
			post_processing_command = 'some cmd7' AND post_processing_environment = '["val here7","another val7"]' AND
			post_processing_client_key = 'aaabbbccc7' AND csr_extra_extensions = '[some csr obj array7]' AND
			preferred_root_cn = 'a root7' AND last_access = 55555557 AND post_processing_client_address = 'example.com/client/';
	`

	rows = db.QueryRowContext(ctx, q)
	count = -1
	err = rows.Scan(&count)
	if err != nil {
		t.Fatalf("validatedatav9: failed to scan certificate 7 (%s)", err)
	}
	if count != 1 {
		t.Errorf("validatedatav9: failed to retrieve certificate 7 (row count expected 1 but got '%d')", count)
	}
}
