package sqlite

import (
	"database/sql"
)

// accountDb is the database representation of an Account object
type accountDb struct {
	id          int
	name        string
	description sql.NullString
	privateKey  *keyDb
	status      sql.NullString
	email       sql.NullString
	acceptedTos sql.NullBool
	isStaging   sql.NullBool
	createdAt   int
	updatedAt   int
	kid         sql.NullString
}

// Turn LE response into Db object
// func acmeAccountToDb(accountId int, response acme_utils.AcmeAccountResponse) accountDb {
// 	var account accountDb

// 	account.id = accountId

// 	// avoid null if there is no contact
// 	account.email.Valid = true
// 	if len(response.Contact) == 0 {
// 		account.email.String = ""
// 	} else {
// 		account.email.String = strings.TrimPrefix(response.Contact[0], "mailto:")
// 	}

// 	unixCreated, err := acme_utils.LeToUnixTime(response.CreatedAt)
// 	if err != nil {
// 		unixCreated = 0
// 	}
// 	account.createdAt = int(unixCreated)

// 	account.updatedAt = int(time.Now().Unix())

// 	account.status.Valid = true
// 	account.status.String = response.Status

// 	account.kid.Valid = true
// 	account.kid.String = response.Location

// 	return account
// }
