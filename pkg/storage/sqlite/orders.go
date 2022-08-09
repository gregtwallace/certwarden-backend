package sqlite

import (
	"database/sql"
)

type orderDb struct {
	id             sql.NullInt32
	acmeAccount    *accountDb
	certificate    *certificateDb
	location       sql.NullString
	status         sql.NullString
	knownRevoked   bool
	err            sql.NullString // stored as json object
	expires        sql.NullInt32
	dnsIdentifiers sql.NullString // will be a comma separated list from storage
	authorizations sql.NullString // will be a comma separated list from storage
	finalize       sql.NullString
	finalizedKey   *keyDb
	certificateUrl sql.NullString
	pem            sql.NullString
	validFrom      sql.NullInt32
	validTo        sql.NullInt32
	createdAt      sql.NullInt32
	updatedAt      sql.NullInt32
}
