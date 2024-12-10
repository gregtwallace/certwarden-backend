package sqlite

import (
	"context"
	"fmt"
)

// CHANGES v6 to v7:
// - certificates:
//     - Add 'preferred_root_cn' field/column
// - acme_orders:
//     - Add 'chain_root_cn' field/column

// migrateV6toV7 updates the storage db from user_version 6 to user_version 7, if it cannot
// do so, an error is returned and modification is aborted
func (store *Storage) migrateV6toV7() (int, error) {
	oldSchemaVer := 6
	newSchemaVer := 7

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

	// add columns
	query = `
		ALTER TABLE certificates ADD preferred_root_cn text NOT NULL DEFAULT "";
	`

	_, err = tx.Exec(query)
	if err != nil {
		return -1, err
	}

	query = `
	ALTER TABLE acme_orders ADD chain_root_cn text;
`

	_, err = tx.Exec(query)
	if err != nil {
		return -1, err
	}

	// update user_version
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
