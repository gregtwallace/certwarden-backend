package sqlite

import (
	"context"
	"legocerthub-backend/pkg/domain/private_keys"
	"legocerthub-backend/pkg/utils"
	"time"
)

// newPayloadToDb translates the new key payload to db object
func newKeyPayloadToDb(payload private_keys.NewPayload) keyDb {
	var dbObj keyDb

	dbObj.name = stringToNullString(payload.Name)

	dbObj.description = stringToNullString(payload.Description)

	dbObj.algorithmValue = stringToNullString(payload.AlgorithmValue)

	dbObj.pem = stringToNullString(payload.PemContent)

	dbObj.createdAt = int(time.Now().Unix())
	dbObj.updatedAt = dbObj.createdAt

	return dbObj
}

// dbPostNewKey creates a new key based on what was POSTed
func (storage *Storage) PostNewKey(payload private_keys.NewPayload) (err error) {
	// load payload fields into db struct
	key := newKeyPayloadToDb(payload)

	// generate api key
	key.apiKey, err = utils.GenerateApiKey()
	if err != nil {
		return err
	}

	// database action
	ctx, cancel := context.WithTimeout(context.Background(), storage.Timeout)
	defer cancel()

	query := `
	INSERT INTO private_keys (name, description, algorithm, pem, api_key, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err = storage.Db.ExecContext(ctx, query,
		key.name,
		key.description,
		key.algorithmValue,
		key.pem,
		key.apiKey,
		key.createdAt,
		key.updatedAt,
	)
	if err != nil {
		return err
	}

	return nil
}
