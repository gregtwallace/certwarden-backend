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
