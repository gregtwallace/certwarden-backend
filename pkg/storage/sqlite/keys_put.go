package sqlite

import (
	"context"
	"legocerthub-backend/pkg/private_keys"
)

// dbPutExistingKey sets an existing key equal to the PUT values (overwriting
//  old values)
func (storage *Storage) PutExistingKey(payload private_keys.KeyPayload) error {
	// load payload fields into db struct
	key, err := payloadToDb(payload)
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
