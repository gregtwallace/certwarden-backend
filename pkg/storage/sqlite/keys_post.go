package sqlite

import (
	"context"
	"legocerthub-backend/pkg/domain/private_keys"
	"legocerthub-backend/pkg/utils"
)

// newPayloadToDb translates the new key payload to db object
func newKeyPayloadToDb(payload private_keys.NewPayload) keyDb {
	var dbObj keyDb

	dbObj.name = stringToNullString(payload.Name)

	dbObj.description = stringToNullString(payload.Description)

	dbObj.algorithmValue = stringToNullString(payload.AlgorithmValue)

	dbObj.pem = stringToNullString(payload.PemContent)

	dbObj.createdAt = timeNow()
	dbObj.updatedAt = dbObj.createdAt

	return dbObj
}

// dbPostNewKey creates a new key based on what was POSTed
func (store *Storage) PostNewKey(payload private_keys.NewPayload) (id int, err error) {
	// load payload fields into db struct
	keyDb := newKeyPayloadToDb(payload)

	// generate api key
	apiKey, err := utils.GenerateApiKey()
	keyDb.apiKey = stringToNullString(&apiKey)
	if err != nil {
		return -2, err
	}

	// database action
	ctx, cancel := context.WithTimeout(context.Background(), store.Timeout)
	defer cancel()

	query := `
	INSERT INTO private_keys (name, description, algorithm, pem, api_key, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	RETURNING id
	`

	// insert and scan the new id
	err = store.Db.QueryRowContext(ctx, query,
		keyDb.name,
		keyDb.description,
		keyDb.algorithmValue,
		keyDb.pem,
		keyDb.apiKey,
		keyDb.createdAt,
		keyDb.updatedAt,
	).Scan(&id)

	if err != nil {
		return -2, err
	}

	return id, nil
}
