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
func newAccountPayloadToDb(payload acme_accounts.NewPayload) (accountDb, error) {
	var dbObj accountDb
	// initialize to avoid nil pointer
	dbObj.privateKey = new(keyDb)

	// mandatory, error if somehow does not exist
	if payload.Name == nil {
		return accountDb{}, errors.New("accounts: new payload: missing name")
	}
	dbObj.name = *payload.Name

	dbObj.description = stringToNullString(payload.Description)

	dbObj.email = stringToNullString(payload.Email)

	// mandatory, error if somehow does not exist
	if payload.PrivateKeyID == nil {
		return accountDb{}, errors.New("accounts: new payload: missing private key id")
	}
	dbObj.privateKey.id = *payload.PrivateKeyID

	dbObj.isStaging = boolToNullBool(payload.IsStaging)

	dbObj.acceptedTos = boolToNullBool(payload.AcceptedTos)

	// initial ACME state is not known until later interaction with ACME
	dbObj.status.Valid = true
	dbObj.status.String = "unknown"

	dbObj.createdAt = int(time.Now().Unix())
	dbObj.updatedAt = dbObj.createdAt

	return dbObj, nil
}

// PostNewAccount inserts a new account into the db
func (storage *Storage) PostNewAccount(payload acme_accounts.NewPayload) (account acme_accounts.Account, err error) {
	// Load payload into db obj
	accountDb, err := newAccountPayloadToDb(payload)
	if err != nil {
		return acme_accounts.Account{}, err
	}

	// database update
	ctx, cancel := context.WithTimeout(context.Background(), storage.Timeout)
	defer cancel()

	tx, err := storage.Db.BeginTx(ctx, nil)
	if err != nil {
		return acme_accounts.Account{}, err
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
	).Scan(&accountDb.id)

	if err != nil {
		tx.Rollback()
		return acme_accounts.Account{}, err
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
		return acme_accounts.Account{}, err
	} else if exists {
		tx.Rollback()
		return acme_accounts.Account{}, errors.New("private key in use by certificate")
	}

	err = tx.Commit()
	if err != nil {
		return acme_accounts.Account{}, err
	}

	account, err = accountDb.accountDbToAcc()
	if err != nil {
		return acme_accounts.Account{}, err
	}

	return account, nil
}
