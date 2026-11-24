package storage

import (
	"context"
	"database/sql"
)

// DeleteCert deletes a cert from the database
func (store *Storage) DeleteCert(id int) (err error) {
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
	temp := -2
	row.Scan(&temp)
	if temp == -2 {
		return sql.ErrNoRows
	}

	// delete
	query = `
	DELETE FROM
		certificates
	WHERE
		id = $1
	`

	_, err = tx.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}
