package sqlite

import (
	"context"
	"fmt"
)

// CHANGES v5 to v6:
// - certificates:
//     - If a certificate with the name `legocerthub` exists, rename it to `serverdefault`
// Note: DB Schema doesn't actually change from 5 to 6.

// migrateV5toV6 updates the storage db from user_version 5 to user_version 6, if it cannot
// do so, an error is returned and modification is aborted
func (store *Storage) migrateV5toV6() (int, error) {
	oldSchemaVer := 5
	newSchemaVer := 6

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

	// if it exists, rename certificate `legocerthub` to `serverdefault`
	query = `
		UPDATE certificates
		SET name = 'serverdefault'
		WHERE name = 'legocerthub'
	`
	_, err = tx.ExecContext(ctx, query)
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
