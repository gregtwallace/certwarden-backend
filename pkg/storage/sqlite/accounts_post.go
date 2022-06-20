package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"legocerthub-backend/pkg/domain/acme_accounts"
)

// postNewAccount inserts a new account into the db and returns the id of the new account
func (storage *Storage) PostNewAccount(payload acme_accounts.AccountPayload) (int, error) {
	// Load payload into db obj
	accountDb, err := accountPayloadToDb(payload)
	if err != nil {
		return 0, err
	}

	// database update
	ctx, cancel := context.WithTimeout(context.Background(), storage.Timeout)
	defer cancel()

	tx, err := storage.Db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}

	// insert the new account
	query := `
	INSERT INTO acme_accounts (name, description, private_key_id, status, email, accepted_tos, is_staging, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	result, err := tx.ExecContext(ctx, query,
		accountDb.name,
		accountDb.description,
		accountDb.privateKeyId,
		accountDb.status,
		accountDb.email,
		accountDb.acceptedTos,
		accountDb.isStaging,
		accountDb.createdAt,
		accountDb.updatedAt,
	)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	// id of the new account
	id, err := result.LastInsertId()
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	// verify the new account does not have a cert that uses the same key
	query = `
		SELECT private_key_id
		FROM
		  certificates
		WHERE
			private_key_id = $1
	`

	row := tx.QueryRowContext(ctx, query, accountDb.privateKeyId)

	var exists bool
	err = row.Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		tx.Rollback()
		return 0, err
	} else if exists {
		tx.Rollback()
		return 0, errors.New("private key in use by certificate")
	}

	err = tx.Commit()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}
