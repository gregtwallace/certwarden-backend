package sqlite

import (
	"database/sql"
)

// certificateDb is the database representation of a certificate object
type certificateDb struct {
	id                   sql.NullInt32
	name                 sql.NullString
	description          sql.NullString
	privateKey           *keyDb
	acmeAccount          *accountDb
	challengeMethodValue sql.NullString
	subject              sql.NullString
	subjectAltNames      sql.NullString // will be a comma separated list from storage
	commonName           sql.NullString
	organization         sql.NullString
	organizationalUnit   sql.NullString
	country              sql.NullString
	state                sql.NullString
	city                 sql.NullString
	createdAt            sql.NullInt32
	updatedAt            sql.NullInt32
	apiKey               sql.NullString
	pem                  sql.NullString
	validFrom            sql.NullInt32
	validTo              sql.NullInt32
}
