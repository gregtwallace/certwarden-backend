package sqlite

import "database/sql"

type orderDb struct {
	id             sql.NullInt32
	acmeAccount    *accountDb
	certificate    *certificateDb
	location       sql.NullString
	status         sql.NullString
	err            sql.NullString // stored as json object
	expires        sql.NullInt32
	dnsIdentifiers sql.NullString // will be a comma separated list from storage
	authorizations sql.NullString // will be a comma separated list from storage
	finalize       sql.NullString
	finalizedKey   *keyDb
	certificateUrl sql.NullString
	createdAt      sql.NullInt32
	updatedAt      sql.NullInt32
}
