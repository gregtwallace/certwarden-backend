package sqlite

import (
	"context"
	"errors"
	"legocerthub-backend/pkg/domain/private_keys"
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
	dbObj.id = intToNullInt32(payload.ID)

	dbObj.name = stringToNullString(payload.Name)

	dbObj.description = stringToNullString(payload.Description)

	dbObj.updatedAt = timeNow()

	return dbObj, nil
}

// dbPutExistingKey sets an existing key equal to the PUT values (overwriting
//  old values)
func (store *Storage) PutNameDescKey(payload private_keys.NameDescPayload) (err error) {
	// load payload fields into db struct
	keyDb, err := nameDescKeyPayloadToDb(payload)
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
		name = $1,
		description = $2,
		updated_at = $3
	WHERE
		id = $4
	`

	_, err = store.Db.ExecContext(ctx, query,
		keyDb.name,
		keyDb.description,
		keyDb.updatedAt,
		keyDb.id)

	if err != nil {
		return err
	}

	return nil
}
