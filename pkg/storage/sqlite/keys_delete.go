package sqlite

import (
	"context"
	"legocerthub-backend/pkg/storage"
)

// DeleteKey deletes a private key from the database
func (store *Storage) DeleteKey(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), store.Timeout)
	defer cancel()

	tx, err := store.Db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// check key exists
	// if scan in succeeds, key exists
	query := `
	SELECT id
	FROM private_keys
	WHERE id = $1
	`

	row := tx.QueryRowContext(ctx, query, id)
	temp := -2
	row.Scan(&temp)
	if temp == -2 {
		return storage.ErrNoRecord
	}

	// check not in use in accounts
	// if scan in succeeds, record exists in acme_accounts
	query = `
	SELECT id
	FROM acme_accounts
	WHERE private_key_id = $1
	`

	row = tx.QueryRowContext(ctx, query, id)
	temp = -2
	row.Scan(&temp)
	if temp != -2 {
		return storage.ErrInUse
	}

	// check not in use in certs
	// if scan in succeeds, record exists in certificates
	query = `
	SELECT id
	FROM certificates
	WHERE private_key_id = $1
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
		private_keys
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
