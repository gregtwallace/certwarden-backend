package sqlite

import (
	"context"
	"errors"
	"legocerthub-backend/pkg/domain/private_keys"
)

// keyInfoPayloadToDb translates the modify key info payload to a db object
func keyInfoPayloadToDb(payload private_keys.InfoPayload) (keyDb, error) {
	var dbObj keyDb
	var err error

	// payload ID should never be missing at this point, regardless error if it somehow
	//  is to avoid nil pointer dereference
	if payload.ID == nil {
		err = errors.New("id missing in payload")
		return keyDb{}, err
	}
	dbObj.id = intToNullInt32(payload.ID)

	dbObj.name = stringToNullString(payload.Name)

	dbObj.description = stringToNullString(payload.Description)

	dbObj.apiKeyViaUrl = *payload.ApiKeyViaUrl

	dbObj.updatedAt = timeNow()

	return dbObj, nil
}

// dbPutExistingKey sets an existing key equal to the PUT values (overwriting
// old values)
func (store *Storage) PutKeyInfo(payload private_keys.InfoPayload) (err error) {
	// load payload fields into db struct
	keyDb, err := keyInfoPayloadToDb(payload)
	if err != nil {
		return err
	}

	// database action
	ctx, cancel := context.WithTimeout(context.Background(), store.Timeout)
	defer cancel()

	query := `
	UPDATE
		private_keys
	SET
		name = case when $1 is null then name else $1 end,
		description = case when $2 is null then description else $2 end,
		api_key_via_url = case when $3 is null then description else $3 end,
		updated_at = $4
	WHERE
		id = $5
	`

	_, err = store.Db.ExecContext(ctx, query,
		keyDb.name,
		keyDb.description,
		keyDb.apiKeyViaUrl,
		keyDb.updatedAt,
		keyDb.id)

	if err != nil {
		return err
	}

	return nil
}
