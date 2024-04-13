package sqlite

import (
	"certwarden-backend/pkg/storage"
	"context"
)

// AccountHasCerts returns true if the specified accountId matches
// any of the certificates in the db
func (store *Storage) AccountHasCerts(accountId int) bool {
	ctx, cancel := context.WithTimeout(context.Background(), store.timeout)
	defer cancel()

	// don't check account exists, business logic in app should do this

	// check account id is not in use in certificates
	query := `
	SELECT id
	FROM certificates
	WHERE acme_account_id = $1
	`

	row := store.db.QueryRowContext(ctx, query, accountId)
	temp := -2

	err := row.Scan(&temp)
	// error means no certs for the account (includes error no rows)
	return err == nil
}

// DeleteAccount deletes an account from the database
func (store *Storage) DeleteAccount(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), store.timeout)
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
		return storage.ErrNoRecord
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
		return storage.ErrInUse
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
