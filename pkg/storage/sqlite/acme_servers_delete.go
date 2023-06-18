package sqlite

import (
	"context"
	"legocerthub-backend/pkg/storage"
)

// ServerHasAccounts returns true if the specified serverId matches
// any of the accounts in the db
func (store *Storage) ServerHasAccounts(serverId int) bool {
	ctx, cancel := context.WithTimeout(context.Background(), store.timeout)
	defer cancel()

	// check server id is not in use in certificates
	query := `
	SELECT id
	FROM acme_accounts
	WHERE acme_server_id = $1
	`

	row := store.db.QueryRowContext(ctx, query, serverId)
	temp := -2

	err := row.Scan(&temp)
	// error means no accounts for the server (includes error no rows)
	return err == nil
}

// DeleteServer deletes an acme server from the database
func (store *Storage) DeleteServer(serverId int) error {
	ctx, cancel := context.WithTimeout(context.Background(), store.timeout)
	defer cancel()

	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// check server exists
	// if scan in succeeds, key exists
	query := `
	SELECT id
	FROM acme_servers
	WHERE id = $1
	`

	row := tx.QueryRowContext(ctx, query, serverId)
	temp := -2
	row.Scan(&temp)
	if temp == -2 {
		return storage.ErrNoRecord
	}

	// check not in use in accounts
	// if scan in succeeds, record exists in accounts
	query = `
	SELECT id
	FROM acme_accounts
	WHERE acme_server_id = $1
	`

	row = tx.QueryRowContext(ctx, query, serverId)
	temp = -2
	row.Scan(&temp)
	if temp != -2 {
		return storage.ErrInUse
	}

	// delete
	query = `
	DELETE FROM
		acme_servers
	WHERE
		id = $1
	`

	_, err = tx.ExecContext(ctx, query, serverId)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}
