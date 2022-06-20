package sqlite

import (
	"database/sql"
	"legocerthub-backend/pkg/domain/private_keys"
	"legocerthub-backend/pkg/utils"
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

// KeyDbToKey translates the db object into the object the key service expects
func (keyDb *keyDb) keyDbToKey() private_keys.Key {
	return private_keys.Key{
		ID:          keyDb.id,
		Name:        keyDb.name.String,
		Description: keyDb.description.String,
		Algorithm:   utils.AlgorithmByValue(keyDb.algorithmValue.String),
		Pem:         keyDb.pem.String,
		ApiKey:      keyDb.apiKey,
		CreatedAt:   keyDb.createdAt,
		UpdatedAt:   keyDb.updatedAt,
	}
}
