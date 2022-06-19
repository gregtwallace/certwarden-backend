package sqlite

import (
	"context"
	"errors"
	"legocerthub-backend/pkg/private_keys"
	"time"
)

// nameDescPayloadToDb translates the modify name/desc payload to a db object
func nameDescKeyPayloadToDb(payload private_keys.NameDescPayload) (keyDb, error) {
	var dbObj keyDb
	var err error

	// payload ID should never be missing at this point, regardless error if it somehow
	//  is to avoid nil pointer dereference
	if payload.ID == nil {
		err = errors.New("id missing in payload")
		return keyDb{}, err
	}
	dbObj.id = *payload.ID

	dbObj.name = stringToNullString(payload.Name)

	dbObj.description = stringToNullString(payload.Description)

	dbObj.updatedAt = int(time.Now().Unix())

	return dbObj, nil
}

// dbPutExistingKey sets an existing key equal to the PUT values (overwriting
//  old values)
func (storage *Storage) PutNameDescKey(payload private_keys.NameDescPayload) error {
	// load payload fields into db struct
	key, err := nameDescKeyPayloadToDb(payload)
	if err != nil {
		return err
	}

	// database action
	ctx, cancel := context.WithTimeout(context.Background(), storage.Timeout)
	defer cancel()

	query := `
	UPDATE
		private_keys
	SET
		name = $1,
		description = $2,
		updated_at = $3
	WHERE
		id = $4`

	_, err = storage.Db.ExecContext(ctx, query,
		key.name,
		key.description,
		key.updatedAt,
		key.id)
	if err != nil {
		return err
	}

	// TODO: Handle 0 rows updated.

	return nil
}
