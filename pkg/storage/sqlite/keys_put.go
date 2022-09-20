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
		api_key_via_url = case when $3 is null then api_key_via_url else $3 end,
		updated_at = $4
	WHERE
		id = $5
	`

	_, err = store.Db.ExecContext(ctx, query,
		payload.Name,
		payload.Description,
		payload.ApiKeyViaUrl,
		payload.UpdatedAt,
		payload.ID)

	if err != nil {
		return err
	}

	return nil
}
