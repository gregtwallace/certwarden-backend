package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"legocerthub-backend/pkg/domain/acme_accounts"
	"time"
)

type NewPayload struct {
	ID           *int    `json:"id"`
	Name         *string `json:"name"`
	Description  *string `json:"description"`
	Email        *string `json:"email"`
	PrivateKeyID *int    `json:"private_key_id"`
	IsStaging    *bool   `json:"is_staging"`
	AcceptedTos  *bool   `json:"accepted_tos"`
}

// accountPayloadToDb turns the client payload into a db object
func newAccountPayloadToDb(payload acme_accounts.NewPayload) accountDb {
	var dbObj accountDb

	dbObj.name = *payload.Name

	dbObj.description = stringToNullString(payload.Description)

	dbObj.email = stringToNullString(payload.Email)

	dbObj.privateKeyId = intToNullInt32(payload.PrivateKeyID)

	dbObj.isStaging = boolToNullBool(payload.IsStaging)

	dbObj.acceptedTos = boolToNullBool(payload.AcceptedTos)

	// initial ACME state is not known until later interaction with ACME
	dbObj.status.Valid = true
	dbObj.status.String = "unknown"

	dbObj.createdAt = int(time.Now().Unix())
	dbObj.updatedAt = dbObj.createdAt

	return dbObj
}

// PostNewAccount inserts a new account into the db and returns the id of the new account
func (storage *Storage) PostNewAccount(payload acme_accounts.NewPayload) (newAccountId int, err error) {
	// Load payload into db obj
	accountDb := newAccountPayloadToDb(payload)
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

	newAccountId = int(id)

	return newAccountId, nil
}
