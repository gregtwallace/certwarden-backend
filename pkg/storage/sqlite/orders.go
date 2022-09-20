package sqlite

import (
	"database/sql"
	"legocerthub-backend/pkg/domain/private_keys"
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
	finalizedKey   *finalizedKeyDb
	certificateUrl sql.NullString
	pem            sql.NullString
	validFrom      sql.NullInt32
	validTo        sql.NullInt32
	createdAt      sql.NullInt32
	updatedAt      sql.NullInt32
}

// temp workaround while refactoring
type finalizedKeyDb struct {
	id   sql.NullInt32
	name sql.NullString
}

func (fkd finalizedKeyDb) toKey() private_keys.Key {
	if !fkd.id.Valid {
		return private_keys.Key{}
	}

	return private_keys.Key{
		ID:   *nullInt32ToInt(fkd.id),
		Name: *nullStringToString(fkd.name),
	}
}
