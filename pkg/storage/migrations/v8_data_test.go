package migrations_test

import (
	"certwarden-backend/pkg/storage/migrations"
	"context"
	"database/sql"
	"testing"
)

func insertDataV8(t *testing.T, db *sql.DB) {
	ctx, cancel := context.WithTimeout(context.Background(), migrations.MigrateDBTimeout)
	defer cancel()

	//
	// private key -- with new attribute
	q := `
		INSERT INTO private_keys
		VALUES
			(6, "key6", "some key desc 6", "alg-6", "pemkeydata6", "apikey61234", "apikeynew62222", 1, 1,
				10000, 999, 2222);
	`

	_, err := db.ExecContext(ctx, q)
	if err != nil {
		t.Fatalf("insertdatav8: failed to insert private key 6 (%s)", err)
	}

	//
	// certificate -- with new attribute
	q = `
		INSERT INTO certificates
		VALUES
			(6, 6, 0, "cert6", "cert desc 6", "subj6.example.com",
				'["alt6","alt16"]', "org6", "ou6", "countr6", "state6", "city6", "[some csr obj array6]", 'a root6',
				"apkey16", "apkey26",	0, "some cmd6", '["val here6","another val6"]', 'aaabbbccc6', 55555555, 123456,
				444446);
	`

	_, err = db.ExecContext(ctx, q)
	if err != nil {
		t.Fatalf("insertdatav8: failed to insert certificate 6 (%s)", err)
	}
}

func validateDataV8(t *testing.T, db *sql.DB) {
	ctx, cancel := context.WithTimeout(context.Background(), migrations.MigrateDBTimeout)
	defer cancel()

	//
	// private keys
	// new attribute has default val
	q := `
		SELECT COUNT(*) FROM private_keys
		WHERE id = 4 AND name = 'key4' AND description = 'some desc 4' AND algorithm = 'an-alg-4' AND
			pem = 'pemdata4' AND api_key = 'apikey1234' AND api_key_new = 'apikeynew2344' AND api_key_disabled = 1 AND
			api_key_via_url = 0 AND created_at = 2224 AND updated_at = 4444 AND last_access = 0;
	`

	rows := db.QueryRowContext(ctx, q)
	count := -1
	err := rows.Scan(&count)
	if err != nil {
		t.Fatalf("validatedatav8: failed to scan private key 4 (%s)", err)
	}
	if count != 1 {
		t.Errorf("validatedatav8: failed to retrieve private key 4 (row count expected 1 but got '%d')", count)
	}

	// new cert with new val
	q = `
		SELECT COUNT(*) FROM private_keys
		WHERE id = 6 AND name = 'key6' AND description = 'some key desc 6' AND algorithm = 'alg-6' AND
			pem = 'pemkeydata6' AND api_key = 'apikey61234' AND api_key_new = 'apikeynew62222' AND api_key_disabled = 1 AND
			api_key_via_url = 1 AND created_at = 999 AND updated_at = 2222 AND last_access = 10000;
	`

	rows = db.QueryRowContext(ctx, q)
	count = -1
	err = rows.Scan(&count)
	if err != nil {
		t.Fatalf("validatedatav8: failed to scan private key 6 (%s)", err)
	}
	if count != 1 {
		t.Errorf("validatedatav8: failed to retrieve private key 6 (row count expected 1 but got '%d')", count)
	}

	//
	// certificates
	// new attribute has default val
	q = `
		SELECT COUNT(*) FROM certificates
		WHERE id = 1 AND private_key_id = 1 AND acme_account_id = 1 AND name = 'some cert 1' AND description = 'cert desc 1' AND 
			subject = 'subj1.example.com' AND subject_alts = '["alt2","alt3"]' AND
			csr_org = 'org1' AND csr_ou = 'ou1' AND csr_country = 'countr1' AND csr_state = 'state1' AND csr_city = 'city1' AND 
			api_key = 'apkey11' AND api_key_new = 'apkey21' AND api_key_via_url = 0 AND created_at = 123451 AND updated_at = 444441 AND
			post_processing_command = '' AND post_processing_environment = '[]' AND post_processing_client_key = '' AND
			csr_extra_extensions = '[]' AND preferred_root_cn = '' AND last_access = 0;
	`

	rows = db.QueryRowContext(ctx, q)
	count = -1
	err = rows.Scan(&count)
	if err != nil {
		t.Fatalf("validatedatav8: failed to scan certificate 1 (%s)", err)
	}
	if count != 1 {
		t.Errorf("validatedatav8: failed to retrieve certificate 1 (row count expected 1 but got '%d')", count)
	}

	// new cert with new val
	q = `
		SELECT COUNT(*) FROM certificates
		WHERE id = 6 AND private_key_id = 6 AND acme_account_id = 0 AND name = 'cert6' AND description = 'cert desc 6' AND
			subject = 'subj6.example.com' AND subject_alts = '["alt6","alt16"]' AND
			csr_org = 'org6' AND csr_ou = 'ou6' AND csr_country = 'countr6' AND csr_state = 'state6' AND csr_city = 'city6' AND
			api_key = 'apkey16' AND api_key_new = 'apkey26' AND api_key_via_url = 0 AND created_at = 123456 AND updated_at = 444446 AND
			post_processing_command = 'some cmd6' AND post_processing_environment = '["val here6","another val6"]' AND
			post_processing_client_key = 'aaabbbccc6' AND csr_extra_extensions = '[some csr obj array6]' AND
			preferred_root_cn = 'a root6' AND last_access = 55555555;
	`

	rows = db.QueryRowContext(ctx, q)
	count = -1
	err = rows.Scan(&count)
	if err != nil {
		t.Fatalf("validatedatav8: failed to scan certificate 6 (%s)", err)
	}
	if count != 1 {
		t.Errorf("validatedatav8: failed to retrieve certificate 6 (row count expected 1 but got '%d')", count)
	}
}
