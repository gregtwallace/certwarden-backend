package sqlite

import (
	"context"
	"fmt"
)

// CHANGES v1 to v2:
// - certificates:
//     - Delete 'challenge_method' field/column

// migrateV1toV2 updates the storage db from user_version 1 to user_version 2, if it cannot
// do so, an error is returned and modification is aborted
func (store *Storage) migrateV1toV2() (int, error) {
	oldSchemaVer := 1
	newSchemaVer := 2

	store.logger.Infof("updating database user_version from %d to %d", oldSchemaVer, newSchemaVer)

	ctx, cancel := context.WithTimeout(context.Background(), store.timeout)
	defer cancel()

	// create sql transaction to roll back in the event an error occurs
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return -1, err
	}
	defer tx.Rollback()

	// verify correct current ver
	query := `PRAGMA user_version`
	row := tx.QueryRowContext(ctx, query)
	fileUserVersion := -1
	err = row.Scan(
		&fileUserVersion,
	)
	if err != nil {
		return -1, err
	}
	if fileUserVersion != 0 {
		return -1, fmt.Errorf("cannot update db schema, current version %d (expected %d)", fileUserVersion, oldSchemaVer)
	}

	// rename old data tables
	err = renameOldDbTables(tx)
	if err != nil {
		return -1, err
	}

	// create new tables
	err = createDBTables(tx)
	if err != nil {
		return -1, err
	}

	// copy data from _old tables to new tables
	query = `
	INSERT INTO users SELECT * FROM users_old;
	INSERT INTO private_keys SELECT * FROM private_keys_old;
	INSERT INTO acme_servers SELECT * FROM acme_servers_old;
	INSERT INTO acme_accounts SELECT * FROM acme_accounts_old;
	`

	_, err = tx.Exec(query)
	if err != nil {
		return -1, err
	}

	// copy data from certificates_old (removing method column/field)
	query = `
	INSERT INTO certificates
	SELECT id, private_key_id, acme_account_id, name, description, subject, subject_alts, csr_org, csr_ou,
		csr_country, csr_state, csr_city, api_key, api_key_new, api_key_via_url, created_at, updated_at
	FROM certificates_old;
	`

	_, err = tx.Exec(query)
	if err != nil {
		return -1, err
	}

	// copy data from remaining table
	query = `
	INSERT INTO acme_orders SELECT * FROM acme_orders_old;
	`

	_, err = tx.Exec(query)
	if err != nil {
		return -1, err
	}

	// remove _old tables
	err = removeOldDbTables(tx)
	if err != nil {
		return -1, err
	}

	// update user_version to 2
	query = fmt.Sprintf(`
		PRAGMA user_version = %d
	`, newSchemaVer)

	_, err = tx.Exec(query)
	if err != nil {
		return -1, err
	}

	// no errors, commit transaction
	err = tx.Commit()
	if err != nil {
		return -1, err
	}

	store.logger.Infof("database user_version successfully upgraded from %d to %d", oldSchemaVer, newSchemaVer)
	return newSchemaVer, nil
}
