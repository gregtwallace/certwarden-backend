package storage

import (
	"context"
	"database/sql"
	"errors"
)

// ServerInUse returns true if the specified serverId matches
// any of the accounts in the db
func (store *Storage) ServerInUse(serverId int) (inUse bool, err error) {
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
	FROM acme_servers
	WHERE id = $1
	`

	row := tx.QueryRowContext(ctx, query, serverId)
	_discardVar := -2
	err = row.Scan(&_discardVar)
	if err != nil {
		// sql.ErrNoRows included here
		return false, err
	}

	// check server id is not in use by any accounts
	query = `
	SELECT id
	FROM acme_accounts
	WHERE acme_server_id = $1
	`

	row = tx.QueryRowContext(ctx, query, serverId)
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

// DeleteServer deletes an acme server from the database
func (store *Storage) DeleteServer(serverId int) error {
	// check that delete is safe
	inUse, err := store.ServerInUse(serverId)
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
		acme_servers
	WHERE
		id = $1
	`

	_, err = store.db.ExecContext(ctx, query, serverId)
	if err != nil {
		return err
	}

	return nil
}
