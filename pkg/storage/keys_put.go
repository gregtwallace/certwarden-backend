package storage

import (
	"certwarden-backend/pkg/domain/private_keys"
	"context"
	"errors"
	"fmt"
)

// PutKeyUpdate updates an existing key in the db using any non-null
// fields specified in the UpdatePayload.
func (store *Storage) PutKeyUpdate(payload private_keys.UpdatePayload) (private_keys.Key, error) {
	// database action
	ctx, cancel := context.WithTimeout(store.shutdownContext, store.timeout)
	defer cancel()

	query := `
	UPDATE
		private_keys
	SET
		name = case when $1 is null then name else $1 end,
		description = case when $2 is null then description else $2 end,
		api_key = case when $3 is null then api_key else $3 end,
		api_key_new = case when $4 is null then api_key_new else $4 end,
		api_key_disabled = case when $5 is null then api_key_disabled else $5 end,
		api_key_via_url = case when $6 is null then api_key_via_url else $6 end,
		updated_at = $7
	WHERE
		id = $8
	`

	res, err := store.db.ExecContext(ctx, query,
		payload.Name,
		payload.Description,
		payload.ApiKey,
		payload.ApiKeyNew,
		payload.ApiKeyDisabled,
		payload.ApiKeyViaUrl,
		payload.UpdatedAt,
		payload.ID,
	)

	if err != nil {
		return private_keys.Key{}, err
	}

	// verify update actually happened
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return private_keys.Key{}, err
	}
	if rowsAffected != 1 {
		return private_keys.Key{}, errors.Join(fmt.Errorf("expected 1 row update, but got '%d'", rowsAffected), ErrWrongUpdateRowCount)
	}

	// get updated key to return
	updatedKey, err := store.GetOneKeyById(payload.ID)
	if err != nil {
		return private_keys.Key{}, err
	}

	return updatedKey, nil
}

// PutKeyApiKey sets a key's api key and updates the updated at time
func (store *Storage) PutKeyApiKey(keyId int, apiKey string, updateTimeUnix int) (err error) {
	// leverage main Put function
	payload := private_keys.UpdatePayload{
		ID:        keyId,
		ApiKey:    &apiKey,
		UpdatedAt: updateTimeUnix,
	}

	_, err = store.PutKeyUpdate(payload)
	return err
}

// PutKeyUpdate sets a key's new api key and updates the updated at time
func (store *Storage) PutKeyNewApiKey(keyId int, newApiKey string, updateTimeUnix int) (err error) {
	// leverage main Put function
	payload := private_keys.UpdatePayload{
		ID:        keyId,
		ApiKeyNew: &newApiKey,
		UpdatedAt: updateTimeUnix,
	}

	_, err = store.PutKeyUpdate(payload)
	return err
}

// PutKeyLastAccess sets a key's last access time
func (store *Storage) PutKeyLastAccess(keyId int, lastAccessTimeUnix int64) (err error) {
	// database action
	ctx, cancel := context.WithTimeout(store.shutdownContext, store.timeout)
	defer cancel()

	query := `
	UPDATE
		private_keys
	SET
		last_access = $1
	WHERE
		id = $2
	`

	res, err := store.db.ExecContext(ctx, query,
		lastAccessTimeUnix,
		keyId,
	)
	if err != nil {
		return err
	}

	// verify update actually happened
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected != 1 {
		return errors.Join(fmt.Errorf("expected 1 row update, but got '%d'", rowsAffected), ErrWrongUpdateRowCount)
	}

	return nil
}
