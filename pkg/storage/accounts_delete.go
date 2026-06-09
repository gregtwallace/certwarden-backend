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

	// check server exists
	query := `
	SELECT id
	FROM acme_accounts
	WHERE id = $1
	`

	row := store.db.QueryRowContext(ctx, query, accountId)
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

	row = store.db.QueryRowContext(ctx, query, accountId)
	err = row.Scan(&_discardVar)
	if !errors.Is(err, sql.ErrNoRows) {
		return true, err
	}

	return false, nil
}

// DeleteAccount deletes an account from the database
func (store *Storage) DeleteAcmeAccount(id int) error {
	ctx, cancel := context.WithTimeout(store.shutdownContext, store.timeout)
	defer cancel()

	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// check acct exists
	// if scan in succeeds, key exists
	query := `
	SELECT id
	FROM acme_accounts
	WHERE id = $1
	`

	row := tx.QueryRowContext(ctx, query, id)
	temp := -2
	row.Scan(&temp)
	if temp == -2 {
		return sql.ErrNoRows
	}

	// check not in use in certs
	// if scan in succeeds, record exists in certificates
	query = `
	SELECT id
	FROM certificates
	WHERE acme_account_id = $1
	`

	row = tx.QueryRowContext(ctx, query, id)
	temp = -2
	row.Scan(&temp)
	if temp != -2 {
		return ErrInUse
	}

	// delete
	query = `
	DELETE FROM
		acme_accounts
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
