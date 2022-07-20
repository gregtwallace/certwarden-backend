package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"legocerthub-backend/pkg/domain/acme_accounts"
)

// accountPayloadToDb turns the client payload into a db object
func newAccountPayloadToDb(payload acme_accounts.NewPayload) (accountDb, error) {
	var dbObj accountDb
	// initialize to avoid nil pointer
	dbObj.privateKey = new(keyDb)

	// mandatory, error if somehow does not exist
	if payload.Name == nil {
		return accountDb{}, errors.New("accounts: new payload: missing name")
	}
	dbObj.name = stringToNullString(payload.Name)

	dbObj.description = stringToNullString(payload.Description)

	dbObj.email = stringToNullString(payload.Email)

	// mandatory, error if somehow does not exist
	if payload.PrivateKeyID == nil {
		return accountDb{}, errors.New("accounts: new payload: missing private key id")
	}
	dbObj.privateKey.id = intToNullInt32(payload.PrivateKeyID)

	dbObj.isStaging = boolToNullBool(payload.IsStaging)

	dbObj.acceptedTos = boolToNullBool(payload.AcceptedTos)

	// initial ACME state is not known until later interaction with ACME
	dbObj.status.Valid = true
	dbObj.status.String = "unknown"

	dbObj.createdAt = timeNow()
	dbObj.updatedAt = dbObj.createdAt

	return dbObj, nil
}

// PostNewAccount inserts a new account into the db
func (store *Storage) PostNewAccount(payload acme_accounts.NewPayload) (id int, err error) {
	// Load payload into db obj
	accountDb, err := newAccountPayloadToDb(payload)
	if err != nil {
		return -2, err
	}

	// database update
	ctx, cancel := context.WithTimeout(context.Background(), store.Timeout)
	defer cancel()

	tx, err := store.Db.BeginTx(ctx, nil)
	if err != nil {
		return -2, err
	}

	// insert the new account
	query := `
	INSERT INTO acme_accounts (name, description, private_key_id, status, email, accepted_tos, is_staging, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	RETURNING id
	`

	err = tx.QueryRowContext(ctx, query,
		accountDb.name,
		accountDb.description,
		accountDb.privateKey.id,
		accountDb.status,
		accountDb.email,
		accountDb.acceptedTos,
		accountDb.isStaging,
		accountDb.createdAt,
		accountDb.updatedAt,
	).Scan(&id)

	if err != nil {
		tx.Rollback()
		return -2, err
	}

	// table already enforces unique private_key_id, so no need to check
	// verify the new account does not have a cert that uses the same key
	query = `
		SELECT private_key_id
		FROM
		  certificates
		WHERE
			private_key_id = $1
	`

	row := tx.QueryRowContext(ctx, query, accountDb.privateKey.id)

	var exists bool
	err = row.Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		tx.Rollback()
		return -2, err
	} else if exists {
		tx.Rollback()
		return -2, errors.New("private key in use by certificate")
	}

	err = tx.Commit()
	if err != nil {
		return -2, err
	}

	return id, nil
}
