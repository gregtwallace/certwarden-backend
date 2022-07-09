package sqlite

import (
	"context"
	"legocerthub-backend/pkg/storage"
)

// DeleteKey deletes a private key from the database
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
