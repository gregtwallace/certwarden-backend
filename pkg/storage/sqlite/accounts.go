package sqlite

import (
	"database/sql"
	"errors"
	"legocerthub-backend/pkg/domain/acme_accounts"
	"legocerthub-backend/pkg/utils/acme_utils"
	"strings"
	"time"
)

// accountDb is the database representation of an Account object
type accountDb struct {
	id             int
	name           string
	description    sql.NullString
	privateKeyId   sql.NullInt32
	privateKeyName sql.NullString // comes from a join with key table
	status         sql.NullString
	email          sql.NullString
	acceptedTos    sql.NullBool
	isStaging      sql.NullBool
	createdAt      int
	updatedAt      int
	kid            sql.NullString
}

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

	dbObj.name = *payload.Name

	dbObj.description = stringToNullString(payload.Description)

	dbObj.email = stringToNullString(payload.Email)

	dbObj.privateKeyId = intToNullInt32(payload.PrivateKeyID)

	// TODO: Cleanup - probably shouldn't need this
	dbObj.status.Valid = true
	dbObj.status.String = "Unknown"

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
