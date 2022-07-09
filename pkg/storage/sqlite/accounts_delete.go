package sqlite

import (
	"context"
	"legocerthub-backend/pkg/storage"
)

// AccountInUse returns a bool if the specified account is in use, it returns
// an error if the account does not exist (or any other error)
func (store *Storage) AccountInUse(id int) (inUse bool, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), store.Timeout)
	defer cancel()

	// check account exists
	// if scan in succeeds, account exists
	query := `
	SELECT id
	FROM acme_accounts
	WHERE id = $1
	`

	row := store.Db.QueryRowContext(ctx, query, id)
	temp := -2
	row.Scan(&temp)
	if temp == -2 {
		return false, storage.ErrNoRecord
	}

	// check not in use in certs
	// if scan in succeeds, record exists in certificates
	query = `
	SELECT id
	FROM certificates
	WHERE acme_account_id = $1
	`

	row = store.Db.QueryRowContext(ctx, query, id)
	temp = -2
	row.Scan(&temp)
	if temp != -2 {
		return true, nil
	}

	return false, nil
}

// DeleteAccount deletes an account from the database
func (store *Storage) DeleteAccount(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), store.Timeout)
	defer cancel()

	tx, err := store.Db.BeginTx(ctx, nil)
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

	_, err = store.Db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}
