package sqlite

import (
	"context"
	"fmt"
)

// CHANGES v8 to v9:
// - certificates:
//     - Add 'post_processing_client_address' field/column

// migrateV8toV9 updates the storage db from user_version 8 to user_version 9, if it cannot
// do so, an error is returned and modification is aborted
func (store *Storage) migrateV8toV9() (int, error) {
	oldSchemaVer := 8
	newSchemaVer := 9

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

	// add column
	query = `
		ALTER TABLE certificates ADD post_processing_client_address text NOT NULL DEFAULT "";
	`

	_, err = tx.Exec(query)
	if err != nil {
		return -1, err
	}

	// populate `post_processing_client_address` if `post_processing_client_key` is not empty
	query = `
		UPDATE certificates
		SET post_processing_client_address = subject
		WHERE post_processing_client_key <> ''
		AND
		subject NOT LIKE '*%'
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
