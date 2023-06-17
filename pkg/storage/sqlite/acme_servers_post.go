package sqlite

import (
	"context"
	"legocerthub-backend/pkg/domain/acme_servers"
)

// PostNewServer saves the KeyExtended to the db as a new key
func (store *Storage) PostNewServer(payload acme_servers.NewPayload) (acmeServerId int, err error) {
	// database action
	ctx, cancel := context.WithTimeout(context.Background(), store.timeout)
	defer cancel()

	query := `
	INSERT INTO acme_servers (name, description, directory_url, is_staging, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6)
	RETURNING id
	`

	// insert and scan the new id
	err = store.db.QueryRowContext(ctx, query,
		payload.Name,
		payload.Description,
		payload.DirectoryURL,
		payload.IsStaging,
		payload.CreatedAt,
		payload.UpdatedAt,
	).Scan(&acmeServerId)

	if err != nil {
		return -2, err
	}

	return acmeServerId, nil
}
