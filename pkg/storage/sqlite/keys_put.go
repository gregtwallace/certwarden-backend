package sqlite

import (
	"context"
	"legocerthub-backend/pkg/domain/private_keys"
)

// PutKeyUpdate updates an existing key in the db using any non-null
// fields specified in the UpdatePayload.
func (store *Storage) PutKeyUpdate(payload private_keys.UpdatePayload) (err error) {
	// database action
	ctx, cancel := context.WithTimeout(context.Background(), store.Timeout)
	defer cancel()

	query := `
	UPDATE
		private_keys
	SET
		name = case when $1 is null then name else $1 end,
		description = case when $2 is null then description else $2 end,
		api_key_disabled = case when $3 is null then description else $3 end,
		api_key_via_url = case when $4 is null then api_key_via_url else $4 end,
		updated_at = $5
	WHERE
		id = $6
	`

	_, err = store.Db.ExecContext(ctx, query,
		payload.Name,
		payload.Description,
		payload.ApiKeyDisabled,
		payload.ApiKeyViaUrl,
		payload.UpdatedAt,
		payload.ID)

	if err != nil {
		return err
	}

	return nil
}

// PutKeyUpdate sets a key's new api key and updates the updated at time
func (store *Storage) PutKeyNewApiKey(keyId int, newApiKey string, updateTimeUnix int) (err error) {
	// database action
	ctx, cancel := context.WithTimeout(context.Background(), store.Timeout)
	defer cancel()

	query := `
	UPDATE
		private_keys
	SET
		api_key_new = $1,
		updated_at = $2
	WHERE
		id = $3
	`

	_, err = store.Db.ExecContext(ctx, query,
		newApiKey,
		updateTimeUnix,
		keyId,
	)

	if err != nil {
		return err
	}

	return nil
}

// PutKeyApiKey sets a key's api key and updates the updated at time
func (store *Storage) PutKeyApiKey(keyId int, apiKey string, updateTimeUnix int) (err error) {
	// database action
	ctx, cancel := context.WithTimeout(context.Background(), store.Timeout)
	defer cancel()

	query := `
	UPDATE
		private_keys
	SET
		api_key = $1,
		updated_at = $2
	WHERE
		id = $3
	`

	_, err = store.Db.ExecContext(ctx, query,
		apiKey,
		updateTimeUnix,
		keyId,
	)

	if err != nil {
		return err
	}

	return nil
}
