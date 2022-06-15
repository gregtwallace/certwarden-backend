package sqlite

import (
	"database/sql"
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
