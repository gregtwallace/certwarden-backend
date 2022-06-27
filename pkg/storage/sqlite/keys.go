package sqlite

import (
	"database/sql"
)

// a single private key, as database table fields
type keyDb struct {
	id             int
	name           sql.NullString
	description    sql.NullString
	algorithmValue sql.NullString
	pem            sql.NullString
	apiKey         string
	createdAt      int
	updatedAt      int
}
