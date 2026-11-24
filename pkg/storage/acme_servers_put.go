package storage

import (
	"certwarden-backend/pkg/domain/acme_servers"
	"context"
)

// PutServerUpdate updates details about an acme Server
func (store *Storage) PutServerUpdate(payload *acme_servers.UpdatePayload) (*acme_servers.Server, error) {
	// database update
	ctx, cancel := context.WithTimeout(store.shutdownContext, store.timeout)
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

	res, err := store.db.ExecContext(ctx, query,
		payload.Name,
		payload.Description,
		payload.DirectoryURL,
		payload.IsStaging,
		payload.UpdatedAt.Unix(),
		payload.ID,
	)
	if err != nil {
		return nil, err
	}

	// verify update actually happened
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}
	if rowsAffected != 1 {
		return nil, errorWrongAffectedRowCount(1, rowsAffected)
	}

	// get updated server to return
	updatedServer, err := store.GetOneServerById(payload.ID)
	if err != nil {
		return nil, err
	}

	return updatedServer, nil
}
