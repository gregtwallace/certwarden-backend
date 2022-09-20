package sqlite

import (
	"database/sql"
)

// accountDb is the database representation of an Account object
type accountDb struct {
	id          sql.NullInt32
	name        sql.NullString
	description sql.NullString
	privateKey  *keyDbExtended
	status      sql.NullString
	email       sql.NullString
	acceptedTos sql.NullBool
	isStaging   sql.NullBool
	createdAt   sql.NullInt32
	updatedAt   sql.NullInt32
	kid         sql.NullString
}
