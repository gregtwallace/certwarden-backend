package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"legocerthub-backend/pkg/acme_accounts"
	"legocerthub-backend/pkg/utils/acme_utils"
	"strings"
	"time"
)

// accountPayloadToDb turns the client payload into a db object
func accountPayloadToDb(payload acme_accounts.AccountPayload) (accountDb, error) {
	var dbObj accountDb
	var err error

	// payload ID should never be missing at this point, regardless error if it somehow
	//  is to avoid nil pointer dereference
	if payload.ID == nil {
		err = errors.New("id missing in payload")
		return accountDb{}, err
	}
	dbObj.id = *payload.ID

	dbObj.name = payload.Name

	dbObj.description.Valid = true
	dbObj.description.String = payload.Description

	dbObj.email.Valid = true
	dbObj.email.String = payload.Email

	dbObj.privateKeyId = intToNullInt32(payload.PrivateKeyID)

	dbObj.acceptedTos = boolToNullBool(payload.AcceptedTos)

	dbObj.isStaging = boolToNullBool(payload.IsStaging)

	// CreatedAt is always populated but only sometimes used
	dbObj.createdAt = int(time.Now().Unix())

	dbObj.updatedAt = dbObj.createdAt

	return dbObj, nil
}

// Turn LE response into Db object
func acmeAccountToDb(accountId int, response acme_utils.AcmeAccountResponse) accountDb {
	var account accountDb

	account.id = accountId

	// avoid null if there is no contact
	account.email.Valid = true
	if len(response.Contact) == 0 {
		account.email.String = ""
	} else {
		account.email.String = strings.TrimPrefix(response.Contact[0], "mailto:")
	}

	unixCreated, err := acme_utils.LeToUnixTime(response.CreatedAt)
	if err != nil {
		unixCreated = 0
	}
	account.createdAt = int(unixCreated)

	account.updatedAt = int(time.Now().Unix())

	account.status.Valid = true
	account.status.String = response.Status

	account.kid.Valid = true
	account.kid.String = response.Location

	return account
}

// boolToNullBool converts a string to a NullBool
// TODO: Probably move into a different package later
func boolToNullBool(b bool) sql.NullBool {
	var nullBool sql.NullBool

	nullBool.Valid = true
	if b == true {
		nullBool.Bool = true
	} else {
		nullBool.Bool = false
	}

	return nullBool
}

// intToNullInt32 converts an *int to a NullInt32
func intToNullInt32(i *int) sql.NullInt32 {
	var nullInt32 sql.NullInt32

	if i == nil {
		nullInt32.Valid = false
	} else {
		nullInt32.Valid = true
		nullInt32.Int32 = int32(*i)
	}

	return nullInt32
}

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

// dbPutExistingAccount overwrites an existing acme account using specified data
// certain fields cannot be updated, per rfc8555
func (storage *Storage) PutExistingAccount(payload acme_accounts.AccountPayload) error {
	// Load payload into db obj
	accountDb, err := accountPayloadToDb(payload)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), storage.Timeout)
	defer cancel()

	query := `
	UPDATE
		acme_accounts
	SET
		name = $1,
		description = $2,
		email = $3,
		accepted_tos = case when $4 is null then accepted_tos else $4 end,
		updated_at = $5
	WHERE
		id = $6`

	_, err = storage.Db.ExecContext(ctx, query,
		accountDb.name,
		accountDb.description,
		accountDb.email,
		accountDb.acceptedTos,
		accountDb.updatedAt,
		accountDb.id)
	if err != nil {
		return err
	}

	// TODO: Handle 0 rows updated.
	return nil
}

// putLEAccountInfo populates an account with data that is returned by LE when
//  an account is POSTed to
func (storage *Storage) PutLEAccountInfo(id int, response acme_utils.AcmeAccountResponse) error {
	// Load id and response into db obj
	var accountDb accountDb
	accountDb = acmeAccountToDb(id, response)

	ctx, cancel := context.WithTimeout(context.Background(), storage.Timeout)
	defer cancel()

	query := `
	UPDATE
		acme_accounts
	SET
		status = $1,
		email = $2,
		created_at = $3,
		updated_at = $4,
		kid = $5
	WHERE
		id = $6`

	_, err := storage.Db.ExecContext(ctx, query,
		accountDb.status,
		accountDb.email,
		accountDb.createdAt,
		accountDb.updatedAt,
		accountDb.kid,
		accountDb.id)
	if err != nil {
		return err
	}

	// TODO: Handle 0 rows updated.
	return nil
}

// putExistingAccountNameDesc only updates the name and desc in the database
// refactor to more generic for anything that can be updated??
func (storage *Storage) PutExistingAccountNameDesc(payload acme_accounts.AccountPayload) error {
	// Load payload into db obj
	accountDb, err := accountPayloadToDb(payload)
	if err != nil {
		return err
	}

	// database update

	ctx, cancel := context.WithTimeout(context.Background(), storage.Timeout)
	defer cancel()

	query := `
	UPDATE
		acme_accounts
	SET
		name = $1,
		description = $2
	WHERE
		id = $3`

	_, err = storage.Db.ExecContext(ctx, query,
		accountDb.name,
		accountDb.description,
		accountDb.id)
	if err != nil {
		return err
	}

	// TODO: Handle 0 rows updated.
	return nil
}
