package sqlite

import (
	"database/sql"
)

// a single private key, as database table fields
type keyDb struct {
	id             sql.NullInt32
	name           sql.NullString
	description    sql.NullString
	algorithmValue sql.NullString
	pem            sql.NullString
	apiKey         sql.NullString
	createdAt      sql.NullInt32
	updatedAt      sql.NullInt32
}
