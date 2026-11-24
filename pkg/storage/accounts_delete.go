package storage

import (
	"context"
	"database/sql"
	"errors"
)

// AcmeAccountInUse returns true if the specified accountId matches
// any of the certificates in the db
func (store *Storage) AcmeAccountInUse(accountId int) (inUse bool, err error) {
	ctx, cancel := context.WithTimeout(store.shutdownContext, store.timeout)
	defer cancel()

	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()

	// check server exists
	query := `
	SELECT id
	FROM acme_accounts
	WHERE id = $1
	`

	row := tx.QueryRowContext(ctx, query, accountId)
	_discardVar := -2
	err = row.Scan(&_discardVar)
	if err != nil {
		// sql.ErrNoRows included here
		return false, err
	}

	// check account id is not in use in certificates
	query = `
	SELECT id
	FROM certificates
	WHERE acme_account_id = $1
	`

	row = tx.QueryRowContext(ctx, query, accountId)
	err = row.Scan(&_discardVar)
	if !errors.Is(err, sql.ErrNoRows) {
		return true, err
	}

	err = tx.Commit()
	if err != nil {
		return false, err
	}

	return false, nil
}

// DeleteAccount deletes an account from the database
func (store *Storage) DeleteAcmeAccount(id int) error {
	// check that delete is safe
	inUse, err := store.AcmeAccountInUse(id)
	if err != nil {
		return err
	}
	if inUse {
		return ErrInUse
	}

	ctx, cancel := context.WithTimeout(store.shutdownContext, store.timeout)
	defer cancel()

	// delete
	query := `
	DELETE FROM
		acme_accounts
	WHERE
		id = $1
	`

	res, err := store.db.ExecContext(ctx, query, id)
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

	return nil
}
