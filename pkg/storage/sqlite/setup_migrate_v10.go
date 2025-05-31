package sqlite

import (
	"context"
	"fmt"
)

// CHANGES v9 to v10:
// - certificates:
//		 - Add 'profile' field/column
// - orders:
//		 - Add 'profile' field/column

// migrateV9toV10 modifies the db to the specified schema, if it cannot
// do so, an error is returned and modification is aborted
func (store *Storage) migrateV9toV10() (int, error) {
	oldSchemaVer := 9
	newSchemaVer := 10

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

	// add profile column to certificates
	query = `
		ALTER TABLE certificates ADD profile text NOT NULL DEFAULT "";
	`

	_, err = tx.Exec(query)
	if err != nil {
		return -1, err
	}

	// add profile column to acme_orders
	query = `
		ALTER TABLE acme_orders ADD profile text DEFAULT NULL;
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
