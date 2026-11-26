package migrations_test

import (
	"certwarden-backend/pkg/storage/migrations"
	"context"
	"database/sql"
	"testing"
)

func insertDataV13(t *testing.T, db *sql.DB) {
	ctx, cancel := context.WithTimeout(context.Background(), migrations.MigrateDBTimeout)
	defer cancel()

	// private key 10 (required by certificates.private_key_id FK)
	q := `
		INSERT INTO private_keys
		VALUES
			(10, "key10", "some desc 10", "an-alg-10", "pemdata10", "apikey12310",
				"apikeynew23410", 1, 1, NULL, 22210, 44410);
	`
	_, err := db.ExecContext(ctx, q)
	if err != nil {
		t.Fatalf("insertdatav13: failed to insert private key 10 (%s)", err)
	}

	q = `
		INSERT INTO certificates
		VALUES
			(10, 10, 0, "cert10", "cert desc 10", "subj10.example.com",
				'["alt10","alt11"]', "org10", "ou10", "countr10", "state10", "city10", "[some csr obj array10]", 'a root10',
				"apkey110", "apkey210", 1, "some cmd10", '["val here10","another val10"]', 'example.com/client/10',
				'aaabbbccc10', 'profile x10', NULL, 123460, 444450, '+351123456789', 'tech@example.com');
	`

	_, err = db.ExecContext(ctx, q)
	if err != nil {
		t.Fatalf("insertdatav13: failed to insert certificate 10 (%s)", err)
	}
}

func validateDataV13(t *testing.T, db *sql.DB) {
	ctx, cancel := context.WithTimeout(context.Background(), migrations.MigrateDBTimeout)
	defer cancel()

	// verify newly inserted cert 10 has tech_phone and tech_email
	q := `
		SELECT COUNT(*) FROM certificates
		WHERE id = 10 AND name = 'cert10' AND tech_phone = '+351123456789' AND tech_email = 'tech@example.com';
	`

	rows := db.QueryRowContext(ctx, q)
	count := -1
	err := rows.Scan(&count)
	if err != nil {
		t.Fatalf("validatedatav13: failed to scan certificate 10 (%s)", err)
	}
	if count != 1 {
		t.Errorf("validatedatav13: failed to retrieve certificate 10 (row count expected 1 but got '%d')", count)
	}

	// verify existing cert 9 from v12 defaulted to empty strings
	q = `
		SELECT COUNT(*) FROM certificates
		WHERE id = 9 AND name = 'cert9' AND tech_phone = '' AND tech_email = '';
	`

	rows = db.QueryRowContext(ctx, q)
	count = -1
	err = rows.Scan(&count)
	if err != nil {
		t.Fatalf("validatedatav13: failed to scan certificate 9 (%s)", err)
	}
	if count != 1 {
		t.Errorf("validatedatav13: failed to retrieve certificate 9 (row count expected 1 but got '%d')", count)
	}
}

