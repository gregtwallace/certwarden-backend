package migrations_test

import (
	"certwarden-backend/pkg/storage/migrations"
	"context"
	"database/sql"
	"testing"
)

func validateDataV6(t *testing.T, db *sql.DB) {
	ctx, cancel := context.WithTimeout(context.Background(), migrations.MigrateDBTimeout)
	defer cancel()

	//
	// certificates
	// verify renamed from 'legocerthub' to 'serverdefault'
	q := `
		SELECT COUNT(*) FROM certificates
		WHERE id = 4 AND private_key_id = 4 AND acme_account_id = 1 AND name = 'serverdefault' AND description = 'cert desc 4' AND
			subject = 'subj4.example.com' AND subject_alts = '["alt4","alt5"]' AND
			csr_org = 'org4' AND csr_ou = 'ou4' AND csr_country = 'countr4' AND csr_state = 'state4' AND csr_city = 'city4' AND
			api_key = 'apkey14' AND api_key_new = 'apkey24' AND api_key_via_url = 0 AND created_at = 123454 AND updated_at = 444444 AND
			post_processing_command = 'some cmd4' AND post_processing_environment = '["val here4","another val4"]' AND
			post_processing_client_key = 'aaabbbccc4' AND csr_extra_extensions = '[some csr obj array4]';
	`

	rows := db.QueryRowContext(ctx, q)
	count := -1
	err := rows.Scan(&count)
	if err != nil {
		t.Fatalf("validatedatav6: failed to scan certificate 4 (%s)", err)
	}
	if count != 1 {
		t.Errorf("validatedatav6: failed to retrieve certificate 4 (row count expected 1 but got '%d')", count)
	}
}
