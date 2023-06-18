package sqlite

import (
	"context"
	"legocerthub-backend/pkg/domain/acme_servers"
)

// PutServerUpdate updates details about an acme Server
func (store *Storage) PutServerUpdate(payload acme_servers.UpdatePayload) (err error) {
	// database update
	ctx, cancel := context.WithTimeout(context.Background(), store.timeout)
	defer cancel()

	query := `
	UPDATE
		acme_servers
	SET
		name = case when $1 is null then name else $1 end,
		description = case when $2 is null then description else $2 end,
		directory_url = case when $3 is null then directory_url else $3 end,
		is_staging = case when $4 is null then is_staging else $4 end,
		updated_at = $5
	WHERE
		id = $6
	`

	_, err = store.db.ExecContext(ctx, query,
		payload.Name,
		payload.Description,
		payload.DirectoryURL,
		payload.IsStaging,
		payload.UpdatedAt,
		payload.ID,
	)

	if err != nil {
		return err
	}

	return nil
}
