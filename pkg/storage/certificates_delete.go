package storage

import (
	"context"
)

// DeleteCert deletes a cert from the database
func (store *Storage) DeleteCert(id int) (err error) {
	// Note: There is no CertInUse func, so this transaction lives here instead

	ctx, cancel := context.WithTimeout(store.shutdownContext, store.timeout)
	defer cancel()

	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// check cert exists
	// if scan in succeeds, cert exists
	query := `
	SELECT id
	FROM certificates
	WHERE id = $1
	`

	row := tx.QueryRowContext(ctx, query, id)
	_discardVar := -2
	err = row.Scan(&_discardVar)
	if err != nil {
		// sql.ErrNoRows included here
		return err
	}

	// delete
	query = `
	DELETE FROM
		certificates
	WHERE
		id = $1
	`

	res, err := tx.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	// verify update actually happened
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected != 1 {
		return errorWrongAffectedRowCount(1, rowsAffected)
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}
