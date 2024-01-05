package sqlite

import (
	"context"
	"fmt"
)

// CHANGES v2 to v3:
// - certificates:
//     - Add 'post_processing_command' field/column
//		 - Add 'post_processing_environment' field/column
//		 - Modify 'subject_alts' from comma separated strings to valid json array object
// - acme_orders:
//		 - Modify 'dns_identifiers' from comma separated strings to valid json array object
//		 - Modify 'authorizations' from comma separated strings to valid json array object

// migrateV2toV3 updates the storage db from user_version 2 to user_version 3, if it cannot
// do so, an error is returned and modification is aborted
func (store *Storage) migrateV2toV3() (int, error) {
	oldSchemaVer := 2
	newSchemaVer := 3

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
		ALTER TABLE certificates ADD post_processing_command text NOT NULL DEFAULT "";
		ALTER TABLE certificates ADD post_processing_environment text NOT NULL DEFAULT "[]"
	`

	_, err = tx.Exec(query)
	if err != nil {
		return -1, err
	}

	// modify column data from comma joined strings to json arrays
	query = `
		UPDATE certificates
			SET subject_alts = case when subject_alts is "" then "[]" else '["' || replace(subject_alts, ',', '","') || '"]' end;

		UPDATE acme_orders
		SET
			dns_identifiers = case when dns_identifiers is "" then "[]" else '["' || replace(dns_identifiers, ',', '","') || '"]' end,
			authorizations = case when authorizations is "" then "[]" else '["' || replace(authorizations, ',', '","') || '"]' end;
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
