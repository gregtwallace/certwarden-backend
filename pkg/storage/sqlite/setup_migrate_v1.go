package sqlite

import (
	"context"
	"fmt"
)

// CHANGES v0 to v1:
// - acme_servers
//     - Add table and fields
//     - Add 2x entries to table for LE Prod (0) and LE Staging (1)
// - acme_accounts:
//     - Add acme_server_id field
//     - Copy is_staging to acme_server_id (matches to 2x entries above without need
//       to change value)
//     - Remove is_staging field

// updates the storage db from user_version 0 to user_version 1, if it cannot
// do so, an error is returned and modification is aborted
func (store *Storage) migrateV0toV1() (int, error) {
	oldSchemaVer := 0
	newSchemaVer := 1

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
	if fileUserVersion != oldSchemaVer {
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

	// add default entries for the new acme_servers table
	err = insertDefaultAcmeServers(tx)
	if err != nil {
		return -1, err
	}

	// copy data from _old tables
	// copy from private_keys
	query = `
	INSERT INTO private_keys SELECT * FROM private_keys_old
	`

	_, err = tx.Exec(query)
	if err != nil {
		return -1, err
	}

	// copy from acme_accounts_old (and transform is_staging to acme_server_id)
	query = `
		INSERT INTO acme_accounts(id, private_key_id, name, description, email, accepted_tos,
			acme_server_id, created_at, updated_at, status, kid)
		SELECT id, private_key_id, name, description, email, accepted_tos, is_staging, created_at,
		  updated_at, status, kid
		FROM acme_accounts_old
	`

	_, err = tx.Exec(query)
	if err != nil {
		return -1, err
	}

	// copy from remaining tables
	query = `
		INSERT INTO certificates SELECT * FROM certificates_old;
		INSERT INTO acme_orders SELECT * FROM acme_orders_old;
		INSERT INTO users SELECT * FROM users_old;
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

	// update user_version to 1
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
