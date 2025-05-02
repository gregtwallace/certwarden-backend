package sqlite

import (
	"context"
	"fmt"
)

// CHANGES v7 to v8:
// - certificates:
//     - Add 'last_access' field/column
// - private_keys:
//     - Add 'last_access' field/column

// migrateV7toV8 updates the storage db from user_version 7 to user_version 8, if it cannot
// do so, an error is returned and modification is aborted
func (store *Storage) migrateV7toV8() (int, error) {
	oldSchemaVer := 7
	newSchemaVer := 8

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
		ALTER TABLE certificates ADD last_access integer NOT NULL DEFAULT 0;
	`

	_, err = tx.Exec(query)
	if err != nil {
		return -1, err
	}

	query = `
		ALTER TABLE private_keys ADD last_access integer NOT NULL DEFAULT 0;
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
