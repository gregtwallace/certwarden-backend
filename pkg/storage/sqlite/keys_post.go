package sqlite

import (
	"context"
	"legocerthub-backend/pkg/domain/private_keys"
)

// newKeyToDb translates a KeyExtended into a db object for storage
func newKeyToDb(newKey private_keys.NewPayload) keyDbExtended {
	var dbObj keyDbExtended

	dbObj.name = *newKey.Name
	dbObj.description = *newKey.Description
	dbObj.algorithmValue = *newKey.AlgorithmValue
	dbObj.pem = *newKey.PemContent
	dbObj.apiKey = newKey.ApiKey
	dbObj.apiKeyViaUrl = newKey.ApiKeyViaUrl
	dbObj.createdAt = newKey.CreatedAt
	dbObj.updatedAt = newKey.UpdatedAt

	return dbObj
}

// PostNewKey saves the KeyExtended to the db as a new key
func (store *Storage) PostNewKey(payload private_keys.NewPayload) (id int, err error) {
	// load payload fields into db struct
	keyDb := newKeyToDb(payload)

	// database action
	ctx, cancel := context.WithTimeout(context.Background(), store.Timeout)
	defer cancel()

	query := `
	INSERT INTO private_keys (name, description, algorithm, pem, api_key, api_key_via_url, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	RETURNING id
	`

	// insert and scan the new id
	err = store.Db.QueryRowContext(ctx, query,
		keyDb.name,
		keyDb.description,
		keyDb.algorithmValue,
		keyDb.pem,
		keyDb.apiKey,
		keyDb.apiKeyViaUrl,
		keyDb.createdAt,
		keyDb.updatedAt,
	).Scan(&id)

	if err != nil {
		return -2, err
	}

	return id, nil
}
